/*
 * Copyright (C) 2020. Nuts community
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 *
 */

package documentlog

import (
	"bytes"
	"errors"
	"fmt"
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/pkg/concurrency"
	"github.com/nuts-foundation/nuts-network/pkg/documentlog/store"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/proto"
	"github.com/nuts-foundation/nuts-network/pkg/stats"
	errors2 "github.com/pkg/errors"
	"io"
	"sync/atomic"
	"time"
)

var AdvertHashInterval = 2 * time.Second

var ErrMissingDocumentContents = errors.New("we don't have the contents for the document (yet)")
var ErrUnknownDocument = errors.New("unknown document")

func NewDocumentLog(protocol proto.Protocol) DocumentLog {
	documentLog := &documentLog{
		protocol:            protocol,
		subscriptions:       make(map[string][]documentQueue, 0),
		subscriptionsMutex:  concurrency.NewSaferRWMutex("doclog-subs"),
		lastConsistencyHash: &atomic.Value{},
	}
	documentLog.lastConsistencyHash.Store(model.EmptyHash())
	return documentLog
}

// documentLog is thread-safe when callers use the DocumentLog interface
type documentLog struct {
	protocol proto.Protocol
	store    store.DocumentStore

	subscriptions       map[string][]documentQueue
	subscriptionsMutex  concurrency.SaferRWMutex
	publicAddr          string
	advertHashTimer     *time.Ticker
	lastConsistencyHash *atomic.Value
}

func (dl *documentLog) Statistics() []stats.Statistic {
	return []stats.Statistic{
		LastConsistencyHashStatistic{Hash: dl.lastConsistencyHash.Load().(model.Hash)},
		NumberOfDocumentsStatistic{NumberOfDocuments: dl.store.Size()},
		LogSizeStatistic{sizeInBytes: dl.store.ContentsSize()},
	}
}

func (dl *documentLog) Configure(store store.DocumentStore) {
	dl.store = store
}

func (dl *documentLog) Subscribe(documentType string) DocumentQueue {
	log.Log().Infof("Creating subscription (document.type=%s)", documentType)
	queue := documentQueue{
		documentType: documentType,
		c:            make(chan *model.Document, 100), // TODO: Does this number make sense?
	}
	dl.subscriptionsMutex.WriteLock(func() {
		if subs, ok := dl.subscriptions[documentType]; ok {
			dl.subscriptions[documentType] = append(subs, queue)
		} else {
			dl.subscriptions[documentType] = []documentQueue{queue}
		}
	})
	return &queue
}

func (dl documentLog) GetDocument(hash model.Hash) (*model.DocumentDescriptor, error) {
	return dl.store.Get(hash)
}

func (dl documentLog) GetDocumentContents(hash model.Hash) (io.ReadCloser, error) {
	contents, err := dl.store.ReadContents(hash)
	if contents == nil {
		return nil, ErrMissingDocumentContents
	}
	return contents, err
}

func (dl *documentLog) AddDocument(document model.Document) error {
	return dl.addDocument(document)
}

func (dl *documentLog) HasContentsForDocument(hash model.Hash) (bool, error) {
	document, err := dl.store.Get(hash)
	if err != nil {
		return false, err
	}
	if document == nil {
		return false, nil
	}
	return document.HasContents, nil
}

func (dl *documentLog) AddDocumentWithContents(timestamp time.Time, documentType string, contents io.Reader) (*model.Document, error) {
	buffer := new(bytes.Buffer)
	if _, err := buffer.ReadFrom(contents); err != nil {
		return nil, err
	}
	document := model.Document{
		Hash:      model.CalculateDocumentHash(documentType, timestamp, buffer.Bytes()),
		Type:      documentType,
		Timestamp: timestamp,
	}
	if err := dl.addDocument(document); err != nil {
		return nil, err
	}
	return dl.AddDocumentContents(document.Hash, buffer)
}

func (dl *documentLog) addDocument(document model.Document) error {
	if document.Timestamp.IsZero() {
		return errors.New("document timestamp is zero")
	}
	existing, err := dl.store.Get(document.Hash)
	if err != nil {
		return err
	}
	if existing != nil {
		// Hash already present, but check if the timestamp matches, just to be sure
		t := existing.Document.Timestamp.UnixNano()
		if t != document.Timestamp.UnixNano() {
			return fmt.Errorf("document hash %s with timestamp %d is already present with different timestamp (%d)", document.Hash, document.Timestamp.UnixNano(), t)
		}
		return nil
	}
	consistencyHash, err := dl.store.Add(document)
	if err != nil {
		return err
	}
	dl.lastConsistencyHash.Store(consistencyHash)
	return nil
}

func (dl *documentLog) AddDocumentContents(hash model.Hash, contents io.Reader) (*model.Document, error) {
	document, err := dl.store.Get(hash)
	if err != nil {
		return nil, err
	} else if document == nil {
		return nil, ErrUnknownDocument
	} else if document.HasContents {
		log.Log().Debugf("Ignoring AddDocumentContents() for document %s since we already have its contents", document.Hash)
		return &document.Document, nil
	}
	if err := dl.store.WriteContents(hash, contents); err != nil {
		return nil, errors2.Wrap(err, "unable to write document contents")
	}
	dl.subscriptionsMutex.ReadLock(func() {
		if subs, ok := dl.subscriptions[document.Type]; ok {
			for _, queue := range subs {
				clone := document.Document.Clone()
				queue.c <- &clone
			}
		}
	})
	return &document.Document, nil
}

func (dl *documentLog) Documents() ([]model.DocumentDescriptor, error) {
	return dl.store.GetAll()
}

func (dl *documentLog) Stop() {
	dl.subscriptionsMutex.WriteLock(func() {
		for _, sub := range dl.subscriptions {
			for _, queue := range sub {
				close(queue.c)
			}
		}
	})
	dl.advertHashTimer.Stop()
}

func (dl *documentLog) Start() {
	dl.advertHashTimer = time.NewTicker(AdvertHashInterval)
	go dl.advertHash()
	go dl.resolveAdvertedHashes(dl.protocol.ReceivedConsistencyHashes())
}

func (dl *documentLog) resolveAdvertedHashes(queue proto.PeerHashQueue) {
	for {
		peerHash := queue.Get()
		if peerHash == nil {
			return
		}
		document, err := dl.store.GetByConsistencyHash(peerHash.Hash)
		if err != nil {
			log.Log().Errorf("Error while checking document (consistency hash=%s) existence: %v", peerHash.Hash, err)
			continue
		}
		if document == nil {
			log.Log().Debugf("Received unknown consistency hash, will query for document hash list (peer=%s,hash=%s)", peerHash.Peer, peerHash.Hash)
			// TODO: Don't have multiple parallel queries for the same peer / hash
			if err := dl.protocol.QueryHashList(peerHash.Peer); err != nil {
				log.Log().Errorf("Could query peer for hash list (peer=%s): %v", peerHash.Peer, err)
			}
		} else {
			log.Log().Tracef("Received known consistency hash, no action is required")
		}
	}
}

func (dl *documentLog) advertHash() {
	for {
		<-dl.advertHashTimer.C
		hash := dl.store.LastConsistencyHash()
		log.Log().Tracef("Adverting last hash (%s)", hash)
		dl.protocol.AdvertConsistencyHash(hash)
		dl.lastConsistencyHash.Store(hash)
	}
}
