package documentlog

import (
	"github.com/nuts-foundation/nuts-go-core"
	log "github.com/nuts-foundation/nuts-network/logging"
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
		subscriptions:       make([]documentQueue, 0),
	}
}

type entry struct {
	// TODO: Cache consistency hash?
	// hash contains the document hash
	hash      model.Hash
	timestamp time.Time
	// doc might not be present when the document isn't resolved (yet)
	doc *model.Document
}

type documentLog struct {
	protocol proto.Protocol
	// entries contains the list of entries in the document log
	entries []entry
	// documentHashIndex contains the entries indexed by document hash
	documentHashIndex map[string]*entry
	// documentHashes contains the list of all document hashes
	documentHashes []model.DocumentHash
	// consistencyHashIndex contains the entries indexed by consistency hash
	consistencyHashIndex map[string]*entry
	lastConsistencyHash  model.Hash
	advertHashTimer      *time.Ticker
	subscriptions        []documentQueue
	publicAddr           string
}

func (dl *documentLog) Diagnostics() []core.DiagnosticResult {
	var sizeInBytes int
	for _, entry := range dl.entries {
		if entry.doc != nil {
			sizeInBytes += len(entry.doc.Contents)
		}
	}
	return []core.DiagnosticResult{
		LastConsistencyHashDiagnostic{Hash: dl.lastConsistencyHash},
		NumberOfDocumentsDiagnostic{NumberOfDocuments: len(dl.entries)},
		LogSizeDiagnostic{SizeInBytes: sizeInBytes},
	}
}

func (dl *documentLog) Configure(publicAddr string) {
	dl.publicAddr = publicAddr
}

type documentQueue struct {
	documentType string
	c            chan *model.Document
}

func (q documentQueue) Get() *model.Document {
	return <-q.c
}

func (dl *documentLog) Subscribe(documentType string) DocumentQueue {
	// TODO: Syncrhonize
	queue := documentQueue{
		documentType: documentType,
		c:            make(chan *model.Document, 100), // TODO: Does this number make sense?
	}
	dl.subscriptions = append(dl.subscriptions, queue)
	return &queue
}

func (dl documentLog) GetDocument(hash model.Hash) *model.Document {
	entry := dl.documentHashIndex[hash.String()]
	if entry == nil {
		return nil
	}
	return entry.doc
}

func (dl *documentLog) HasDocument(hash model.Hash) bool {
	return dl.documentHashIndex[hash.String()].doc != nil
}

func (dl *documentLog) AddDocumentHash(hash model.Hash, timestamp time.Time) {
	if dl.documentHashIndex[hash.String()] != nil {
		// Hash already present, but check if the timestamp matches, just to be sure
		t := dl.documentHashIndex[hash.String()].timestamp.UnixNano()
		if t != timestamp.UnixNano() {
			log.Log().Warnf("Integrity violation! Hash %s with timestamp %d is already present with different timestamp (%d)", hash, timestamp.UnixNano(), hash)
		}
		return
	}
	newEntry := entry{
		hash:      hash,
		timestamp: timestamp,
	}
	dl.entries = append(dl.entries, newEntry)
	dl.documentHashIndex[newEntry.hash.String()] = &newEntry
	// TODO: Isn't there a faster way to keep it sorted (or not sort it at all?)
	// TODO: Synchronization!
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
	for i = 0; i < len(dl.entries); i++ {
		documentHashes[i] = model.DocumentHash{
			Hash:      dl.entries[i].hash,
			Timestamp: dl.entries[i].timestamp,
		}
		if i == 0 {
			copy(prevHash, dl.entries[i].hash)
		} else {
			model.MakeConsistencyHash(prevHash, prevHash, dl.entries[i].hash)
		}
		dl.consistencyHashIndex[prevHash.String()] = &dl.entries[i]
	}
	dl.lastConsistencyHash = prevHash
	dl.documentHashes = documentHashes
}

func (dl *documentLog) AddDocument(document *model.Document) {
	dl.AddDocumentHash(document.Hash(), document.Timestamp)

	entry := dl.documentHashIndex[document.Hash().String()]
	if entry.doc == nil {
		entry.doc = document
		// TODO: Synchronization
		for _, sub := range dl.subscriptions {
			if sub.documentType == document.Type {
				sub.c <- document
			}
		}
	}
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
		if dl.consistencyHashIndex[peerHash.Hash.String()] == nil {
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
			log.Log().Debugf("Adverting last hash (%s)", dl.lastConsistencyHash)
			dl.protocol.AdvertConsistencyHash(dl.lastConsistencyHash)
		}
	}
}
