package documentlog

import (
	"bytes"
	"errors"
	"github.com/nuts-foundation/nuts-go-core"
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/pkg/concurrency"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/proto"
	"io"
	"sort"
	"time"
)

var AdvertHashInterval = 2 * time.Second

var ErrMissingDocumentContents = errors.New("we don't have the contents for the document (yet)")
var ErrUnknownDocument = errors.New("unknown document")

func NewDocumentLog(protocol proto.Protocol) DocumentLog {
	return &documentLog{
		protocol:            protocol,
		lastConsistencyHash: model.EmptyHash(),
		documents:           make([]model.Document, 0),
		documentHashIndex:   make(map[string]*entry, 0),
		documentsMutex:      concurrency.NewSaferRWMutex("doclog-docs"),
		subscriptions:       make([]documentQueue, 0),
		subscriptionsMutex:  concurrency.NewSaferRWMutex("doclog-subs"),
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
	// documents contains the list of all documents
	documents []model.Document
	// consistencyHashIndex contains the entries indexed by consistency hash
	consistencyHashIndex map[string]*entry
	documentsMutex       concurrency.SaferRWMutex
	lastConsistencyHash  model.Hash
	advertHashTimer      *time.Ticker
	subscriptions        []documentQueue
	subscriptionsMutex   concurrency.SaferRWMutex
	publicAddr           string

	// keep diagnostic state separate from source data (share-nothing) to avoid concurrent access
	logSizeDiagnostic             LogSizeDiagnostic
	numberOfDocumentsDiagnostic   NumberOfDocumentsDiagnostic
	lastConsistencyHashDiagnostic LastConsistencyHashDiagnostic
	consistencyHashListDiagnostic ConsistencyHashListDiagnostic
}

func (dl *documentLog) Diagnostics() []core.DiagnosticResult {
	return []core.DiagnosticResult{
		dl.lastConsistencyHashDiagnostic,
		dl.consistencyHashListDiagnostic,
		dl.numberOfDocumentsDiagnostic,
		dl.logSizeDiagnostic,
	}
}

func (dl *documentLog) Configure(publicAddr string) {
	dl.publicAddr = publicAddr
}

func (dl *documentLog) Subscribe(documentType string) DocumentQueue {
	queue := documentQueue{
		documentType: documentType,
		c:            make(chan model.Document, 100), // TODO: Does this number make sense?
	}
	dl.subscriptionsMutex.WriteLock(func() {
		dl.subscriptions = append(dl.subscriptions, queue)
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
	// TODO: Implement this without a lock
	var result bool
	dl.documentsMutex.ReadLock(func() {
		result = dl.documentHashIndex[hash.String()] != nil
	})
	return result
}

func (dl *documentLog) HasContentsForDocument(hash model.Hash) bool {
	// TODO: Implement this without a lock
	var result bool
	dl.documentsMutex.ReadLock(func() {
		result = dl.documentHashIndex[hash.String()] != nil && dl.documentHashIndex[hash.String()].contents != nil
	})
	return result
}

func (dl *documentLog) AddDocumentWithContents(document model.Document, contents io.Reader) error {
	dl.AddDocument(document)
	return dl.AddDocumentContents(document.Hash, contents)
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
		dl.numberOfDocumentsDiagnostic.NumberOfDocuments++
		dl.entries = append(dl.entries, newEntry)
		dl.documentHashIndex[newEntry.Hash.String()] = newEntry
		// TODO: Isn't there a faster way to keep it sorted (or not sort it at all?)
		// TODO: What if entries have the same timestamp? Use hash for ordering
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
		dl.lastConsistencyHash = prevHash
		dl.lastConsistencyHashDiagnostic.Hash = dl.lastConsistencyHash.String()
		dl.consistencyHashListDiagnostic.Hashes = consistencyHashes
		dl.documents = documents
	})
}

func (dl *documentLog) AddDocumentContents(hash model.Hash, contents io.Reader) error {
	var err error
	dl.documentsMutex.WriteLock(func() {
		entry := dl.documentHashIndex[hash.String()]
		if entry == nil {
			err = ErrUnknownDocument
			return
		}
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
			dl.logSizeDiagnostic.SizeInBytes += len(bytes)
			dl.subscriptionsMutex.ReadLock(func() {
				for _, sub := range dl.subscriptions {
					if sub.documentType == entry.Type {
						sub.c <- entry.Document
					}
				}
			})
		}
	})
	return err
}

func (dl *documentLog) Documents() []model.Document {
	return dl.documents
}

func (dl *documentLog) Stop() {
	// TODO: Should check result of Stop()
	dl.advertHashTimer.Stop()
}

func (dl *documentLog) Start() {
	dl.advertHashTimer = time.NewTicker(AdvertHashInterval)
	go dl.advertHash()
	go dl.resolveAdvertedHashes()
}

// TODO: Comment
// resolveAdvertedHashes reads
func (dl *documentLog) resolveAdvertedHashes() {
	// TODO: When to quite the loop?
	queue := dl.protocol.ReceivedConsistencyHashes()
	for {
		peerHash := queue.Get()
		log.Log().Debugf("Got consistency hash (ours: %s, received: %s)", dl.lastConsistencyHash.String(), peerHash.Hash.String())
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
		if !dl.lastConsistencyHash.Empty() {
			log.Log().Debugf("Adverting last hash (%s)", dl.lastConsistencyHash)
			dl.protocol.AdvertConsistencyHash(dl.lastConsistencyHash)
		}
	}
}

// NoopCloser implements io.ReadCloser with a No-Operation, intended for returning byte slices.
type NoopCloser struct {
	io.Reader
	io.Closer
}

func (NoopCloser) Close() error { return nil }
