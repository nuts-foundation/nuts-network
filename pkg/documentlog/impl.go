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
	"context"
	"errors"
	"fmt"
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/pkg/concurrency"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/proto"
	"github.com/nuts-foundation/nuts-network/pkg/stats"
	"io"
	"sort"
	"sync/atomic"
	"time"
)

var AdvertHashInterval = 2 * time.Second

var ErrMissingDocumentContents = errors.New("we don't have the contents for the document (yet)")
var ErrUnknownDocument = errors.New("unknown document")

func NewDocumentLog(protocol proto.Protocol) DocumentLog {
	return &documentLog{
		protocol:            protocol,
		documentHashIndex:   make(map[string]*entry, 0),
		documentsMutex:      concurrency.NewSaferRWMutex("doclog-docs"),
		subscriptions:       make(map[string]documentQueue, 0),
		subscriptionsMutex:  concurrency.NewSaferRWMutex("doclog-subs"),
		statisticsMutex:     concurrency.NewSaferRWMutex("doclog-stats"),
		lastConsistencyHash: &atomic.Value{},
	}
}

type entry struct {
	model.Document
	contents *[]byte
}

// documentLog is thread-safe when callers use the DocumentLog interface
type documentLog struct {
	protocol proto.Protocol
	// entries contains the list of entries in the document log
	entries []*entry
	// documentHashIndex contains the entries indexed by document hash
	documentHashIndex map[string]*entry
	// consistencyHashIndex contains the entries indexed by consistency hash
	consistencyHashIndex map[string]*entry
	documentsMutex       concurrency.SaferRWMutex
	lastConsistencyHash  *atomic.Value
	advertHashTimer      *time.Ticker
	subscriptions        map[string]documentQueue
	subscriptionsMutex   concurrency.SaferRWMutex
	publicAddr           string
	cxt                  context.Context
	cxtCancel            context.CancelFunc

	// keep statistic state separate from source data (share-nothing) to avoid concurrent access
	logSizeStatistics            LogSizeStatistic
	numberOfDocumentsStatistic   NumberOfDocumentsStatistic
	lastConsistencyHashStatistic LastConsistencyHashStatistic
	consistencyHashListStatistic ConsistencyHashListStatistic
	statisticsMutex              concurrency.SaferRWMutex
}

func (dl *documentLog) Statistics() []stats.Statistic {
	var results []stats.Statistic
	dl.statisticsMutex.ReadLock(func() {
		results = []stats.Statistic{
			dl.lastConsistencyHashStatistic,
			dl.consistencyHashListStatistic,
			dl.numberOfDocumentsStatistic,
			dl.logSizeStatistics,
		}
	})
	return results
}

func (dl *documentLog) Configure(publicAddr string) {
	dl.publicAddr = publicAddr
}

func (dl *documentLog) Subscribe(documentType string) DocumentQueue {
	queue := documentQueue{documentType: documentType}
	queue.internal.Init(100) // TODO: Does this number make sense?
	dl.subscriptionsMutex.WriteLock(func() {
		dl.subscriptions[documentType] = queue
	})
	return &queue
}

func (dl documentLog) GetDocument(hash model.Hash) *model.Document {
	var entry *entry
	dl.documentsMutex.ReadLock(func() {
		entry = dl.documentHashIndex[hash.String()]
	})
	if entry == nil {
		return nil
	}
	return &entry.Document
}

func (dl documentLog) GetDocumentContents(hash model.Hash) (io.ReadCloser, error) {
	var entry *entry
	dl.documentsMutex.ReadLock(func() {
		entry = dl.documentHashIndex[hash.String()]
	})
	if entry == nil {
		return nil, ErrUnknownDocument
	}
	// TODO: Nil/len check on Contents?
	if entry.contents == nil {
		return nil, ErrMissingDocumentContents
	}
	return NoopCloser{Reader: bytes.NewReader(*entry.contents)}, nil
}

func (dl *documentLog) HasDocument(hash model.Hash) bool {
	var result bool
	dl.documentsMutex.ReadLock(func() {
		result = dl.documentHashIndex[hash.String()] != nil
	})
	return result
}

func (dl *documentLog) HasContentsForDocument(hash model.Hash) bool {
	var result bool
	dl.documentsMutex.ReadLock(func() {
		result = dl.documentHashIndex[hash.String()] != nil && dl.documentHashIndex[hash.String()].contents != nil
	})
	return result
}

func (dl *documentLog) AddDocumentWithContents(timestamp time.Time, documentType string, contents io.Reader) (model.Document, error) {
	buffer := new(bytes.Buffer)
	if _, err := buffer.ReadFrom(contents); err != nil {
		return model.Document{}, err
	}
	document := model.Document{
		Hash:      model.CalculateDocumentHash(documentType, timestamp, buffer.Bytes()),
		Type:      documentType,
		Timestamp: timestamp,
	}
	if dl.HasContentsForDocument(document.Hash) {
		return model.Document{}, fmt.Errorf("document already exists (with content) for hash: %s", document.Hash)
	}
	dl.AddDocument(document)
	return dl.AddDocumentContents(document.Hash, buffer)
}

func (dl *documentLog) AddDocument(document model.Document) {
	dl.documentsMutex.WriteLock(func() {
		existing := dl.documentHashIndex[document.Hash.String()]
		if existing != nil {
			// Hash already present, but check if the timestamp matches, just to be sure
			t := existing.Document.Timestamp.UnixNano()
			if t != document.Timestamp.UnixNano() {
				log.Log().Warnf("Integrity violation! Document hash %s with timestamp %d is already present with different timestamp (%d)", document.Hash, document.Timestamp.UnixNano(), t)
			}
			return
		}
		newEntry := &entry{Document: document}
		dl.entries = append(dl.entries, newEntry)
		dl.documentHashIndex[newEntry.Hash.String()] = newEntry
		// TODO: Isn't there a faster way to keep it sorted (or not sort it at all?)
		sort.Slice(dl.entries, func(i, j int) bool {
			if dl.entries[i].Timestamp == dl.entries[j].Timestamp {
				return dl.entries[i].Timestamp.Before(dl.entries[j].Timestamp)
			} else {
				return dl.entries[i].Hash.String() < dl.entries[j].Hash.String()
			}
		})
		// Calc last consistency hash
		// TODO: Test this
		// TODO: Make this smarter (retain unchanged consistency hashes)
		dl.consistencyHashIndex = make(map[string]*entry, len(dl.entries))
		documents := make([]model.Document, len(dl.entries))
		var i = 0
		prevHash := model.EmptyHash()
		consistencyHashes := make([]string, len(dl.entries))
		for i = 0; i < len(dl.entries); i++ {
			documents[i] = dl.entries[i].Document
			if i == 0 {
				copy(prevHash, dl.entries[i].Hash)
			} else {
				prevHash = model.MakeConsistencyHash(prevHash, dl.entries[i].Hash)
			}
			consistencyHashes[i] = prevHash.String()
			dl.consistencyHashIndex[prevHash.String()] = dl.entries[i]
		}
		dl.lastConsistencyHash.Store(prevHash)
		dl.statisticsMutex.WriteLock(func() {
			dl.numberOfDocumentsStatistic.NumberOfDocuments++
			dl.lastConsistencyHashStatistic.Hash = prevHash.String()
			dl.consistencyHashListStatistic.Hashes = consistencyHashes
		})
	})
}

func (dl *documentLog) AddDocumentContents(hash model.Hash, contents io.Reader) (model.Document, error) {
	var err error
	var document model.Document
	dl.documentsMutex.WriteLock(func() {
		entry := dl.documentHashIndex[hash.String()]
		if entry == nil {
			err = ErrUnknownDocument
			return
		}
		document = entry.Document
		if entry.contents == nil {
			buffer := new(bytes.Buffer)
			_, err = buffer.ReadFrom(contents)
			if err != nil {
				return
			}
			bytes := buffer.Bytes()
			actualHash := model.CalculateDocumentHash(entry.Type, entry.Timestamp, bytes)
			if !entry.Hash.Equals(hash) {
				log.Log().Warnf("Document rejected, actual hash differs from expected (expected=%s,actual=%s)", entry.Hash, actualHash)
				return
			}
			entry.contents = &bytes
			dl.statisticsMutex.WriteLock(func() {
				dl.logSizeStatistics.sizeInBytes += len(bytes)
			})
			dl.subscriptionsMutex.ReadLock(func() {
				if queue, ok := dl.subscriptions[entry.Type]; ok {
					queue.internal.Add(entry.Document)
				}
			})
		}
	})
	return document.Clone(), err
}

func (dl *documentLog) Documents() []model.Document {
	var result []model.Document
	dl.documentsMutex.ReadLock(func() {
		for _, entry := range dl.entries {
			result = append(result, entry.Document.Clone())
		}
	})
	return result
}

func (dl *documentLog) Stop() {
	// TODO: Should check result of Stop()
	dl.advertHashTimer.Stop()
}

func (dl *documentLog) Start() {
	dl.cxt, dl.cxtCancel = context.WithCancel(context.Background())
	dl.advertHashTimer = time.NewTicker(AdvertHashInterval)
	go dl.advertHash()
	go dl.resolveAdvertedHashes(dl.protocol.ReceivedConsistencyHashes())
}

func (dl *documentLog) resolveAdvertedHashes(queue proto.PeerHashQueue) {
	for {
		peerHash, err := queue.Get(dl.cxt)
		if err != nil {
			log.Log().Debugf("Get cancelled: %v", err)
			return
		}
		log.Log().Debugf("Got consistency hash: %s", peerHash.Hash.String())
		var entry *entry
		dl.documentsMutex.ReadLock(func() {
			entry = dl.consistencyHashIndex[peerHash.Hash.String()]
		})
		if entry == nil {
			log.Log().Debugf("Received unknown consistency hash, will query for document hash list (peer=%s,hash=%s)", peerHash.Peer, peerHash.Hash)
			// TODO: Don't have multiple parallel queries for the same peer / hash
			if err := dl.protocol.QueryHashList(peerHash.Peer); err != nil {
				log.Log().Errorf("Could query peer for hash list (peer=%s): %v", peerHash.Peer, err)
			}
		} else {
			log.Log().Debugf("Received known consistency hash, no action is required")
		}
	}
}

func (dl *documentLog) advertHash() {
	for {
		<-dl.advertHashTimer.C
		hash, ok := dl.lastConsistencyHash.Load().(model.Hash)
		if ok && !hash.Empty() {
			log.Log().Debugf("Adverting last hash (%s)", hash)
			dl.protocol.AdvertConsistencyHash(hash)
		} else if !ok {
			log.Log().Error("lastConsistencyHash doesn't contain instance of model.Hash!")
		}
	}
}

// NoopCloser implements io.ReadCloser with a No-Operation, intended for returning byte slices.
type NoopCloser struct {
	io.Reader
	io.Closer
}

func (NoopCloser) Close() error { return nil }
