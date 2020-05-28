package documentlog

import (
	"github.com/nuts-foundation/nuts-go-core"
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/pkg/concurrency"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/proto"
	"sort"
	"time"
)

const AdvertHashInterval = 2 * time.Second

func NewDocumentLog(protocol proto.Protocol) DocumentLog {
	return &documentLog{
		protocol:            protocol,
		lastConsistencyHash: model.EmptyHash(),
		documentHashes:      make([]model.DocumentHash, 0),
		documentHashIndex:   make(map[string]*entry, 0),
		documentsMutex:      &concurrency.SaferRWMutex{},
		subscriptions:       make([]documentQueue, 0),
		subscriptionsMutex:  &concurrency.SaferRWMutex{},
	}
}

type entry struct {
	// hash contains the document hash
	hash      model.Hash
	timestamp time.Time
	// doc might not be present when the document isn't resolved (yet)
	doc *model.Document
}

// documentLog is thread-safe when callers use the DocumentLog interface
type documentLog struct {
	protocol proto.Protocol
	// entries contains the list of entries in the document log
	entries []*entry
	// documentHashIndex contains the entries indexed by document hash
	documentHashIndex map[string]*entry
	// documentHashes contains the list of all document hashes
	documentHashes []model.DocumentHash
	// consistencyHashIndex contains the entries indexed by consistency hash
	consistencyHashIndex map[string]*entry
	documentsMutex       *concurrency.SaferRWMutex
	lastConsistencyHash  model.Hash
	advertHashTimer      *time.Ticker
	subscriptions        []documentQueue
	subscriptionsMutex   *concurrency.SaferRWMutex
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
		c:            make(chan *model.Document, 100), // TODO: Does this number make sense?
	}
	dl.subscriptionsMutex.WriteLock(func() {
		dl.subscriptions = append(dl.subscriptions, queue)
	})
	return &queue
}

func (dl documentLog) GetDocument(hash model.Hash) model.Document {
	var entry *entry
	dl.documentsMutex.ReadLock(func() {
		entry = dl.documentHashIndex[hash.String()]
	})
	if entry == nil || entry.doc == nil {
		return model.Document{}
	}
	return *entry.doc
}

func (dl *documentLog) HasDocument(hash model.Hash) bool {
	var result bool
	dl.documentsMutex.ReadLock(func() {
		result = dl.documentHashIndex[hash.String()].doc != nil
	})
	return result
}

func (dl *documentLog) AddDocumentHash(hash model.Hash, timestamp time.Time) {
	dl.documentsMutex.WriteLock(func() {
		if dl.documentHashIndex[hash.String()] != nil {
			// Hash already present, but check if the timestamp matches, just to be sure
			t := dl.documentHashIndex[hash.String()].timestamp.UnixNano()
			if t != timestamp.UnixNano() {
				log.Log().Warnf("Integrity violation! Document hash %s with timestamp %d is already present with different timestamp (%d)", hash, timestamp.UnixNano(), hash)
			}
			return
		}
		newEntry := &entry{
			hash:      hash,
			timestamp: timestamp,
		}
		dl.numberOfDocumentsDiagnostic.NumberOfDocuments++
		dl.entries = append(dl.entries, newEntry)
		dl.documentHashIndex[newEntry.hash.String()] = newEntry
		// TODO: Isn't there a faster way to keep it sorted (or not sort it at all?)
		// TODO: What if entries have the same timestamp? Use hash for ordering
		sort.Slice(dl.entries, func(i, j int) bool {
			return dl.entries[i].timestamp.Before(dl.entries[j].timestamp)
		})
		// Calc last consistency hash
		// TODO: Test this
		// TODO: Make this smarter (retain unchanged consistency hashes)
		dl.consistencyHashIndex = make(map[string]*entry, len(dl.entries))
		documentHashes := make([]model.DocumentHash, len(dl.entries))
		var i = 0
		prevHash := model.EmptyHash()
		consistencyHashes := make([]string, len(dl.entries))
		for i = 0; i < len(dl.entries); i++ {
			documentHashes[i] = model.DocumentHash{
				Hash:      dl.entries[i].hash,
				Timestamp: dl.entries[i].timestamp,
			}
			if i == 0 {
				copy(prevHash, dl.entries[i].hash)
			} else {
				prevHash = model.MakeConsistencyHash(prevHash, dl.entries[i].hash)
			}
			consistencyHashes[i] = prevHash.String()
			dl.consistencyHashIndex[prevHash.String()] = dl.entries[i]
		}
		dl.lastConsistencyHash = prevHash
		dl.lastConsistencyHashDiagnostic.Hash = dl.lastConsistencyHash.String()
		dl.consistencyHashListDiagnostic.Hashes = consistencyHashes
		dl.documentHashes = documentHashes
	})
}

func (dl *documentLog) AddDocument(document *model.Document) {
	dl.AddDocumentHash(document.Hash(), document.Timestamp)

	dl.documentsMutex.WriteLock(func() {
		entry := dl.documentHashIndex[document.Hash().String()]
		if entry.doc == nil {
			if !entry.hash.Equals(document.Hash()) {
				log.Log().Warnf("Document rejected, actual hash differs from expected (expected=%s,actual=%s)", entry.hash, document.Hash())
				return
			}
			entry.doc = document
			dl.logSizeDiagnostic.SizeInBytes += len(entry.doc.Contents)
			dl.subscriptionsMutex.ReadLock(func() {
				for _, sub := range dl.subscriptions {
					if sub.documentType == document.Type {
						sub.c <- document
					}
				}
			})
		}
	})
}

func (dl *documentLog) DocumentHashes() []model.DocumentHash {
	return dl.documentHashes
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
				log.Log().Errorf("Could query peer for hash list (peer=%s)", peerHash.Peer, err)
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
			log.Log().Infof("Adverting last hash (%s)", dl.lastConsistencyHash)
			dl.protocol.AdvertConsistencyHash(dl.lastConsistencyHash)
		}
	}
}
