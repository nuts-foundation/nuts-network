package doclog

import (
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/proto"
	"sort"
	"time"
)

const AdvertHashInterval = 2 * time.Second

func NewDocumentLog(protocol proto.Protocol) DocumentLog {
	return &documentLog{
		protocol:          protocol,
		hash:              model.EmptyHash(),
		consistencyHashes: make([]model.Hash, 0),
		entryIndex:        make(map[string]*entry, 0),
	}
}

type entry struct {
	// TODO: Cache consistency hash
	// hash contains the (cached) document hash
	hash model.Hash
	doc  *model.Document
}

type documentLog struct {
	protocol             proto.Protocol
	entries              []entry
	entryIndex           map[string]*entry
	consistencyHashes    []model.Hash
	consistencyHashIndex map[string]*entry
	hash                 model.Hash
	advertHashTimer      *time.Ticker
}

func (dl *documentLog) ContainsDocument(hash model.Hash) bool {
	return dl.entryIndex[hash.String()] != nil
}

func (dl *documentLog) Add(document *model.Document) {
	entry := entry{
		hash: document.Hash(),
		doc:  document,
	}
	dl.entries = append(dl.entries, entry)
	dl.entryIndex[entry.hash.String()] = &entry
	// TODO: Isn't there a faster way to keep it sorted (or not sort it at all?)
	// TODO: Synchronization!
	sort.Slice(dl.entries, func(i, j int) bool {
		return dl.entries[i].doc.Timestamp.Before(dl.entries[j].doc.Timestamp)
	})
	// Calc last consistency hash
	// TODO: Test this
	// TODO: Make this smarter (retain unchanged consistency hashes)
	dl.consistencyHashIndex = make(map[string]*entry, len(dl.entries))
	consistencyHashes := make([]model.Hash, len(dl.entries))
	var i = 0
	prevHash := model.EmptyHash()
	for i = 0; i < len(dl.entries); i++ {
		if i == 0 {
			copy(prevHash, dl.entries[i].hash)
		} else {
			model.MakeConsistencyHash(prevHash, prevHash, dl.entries[i].hash)
		}
		dl.consistencyHashIndex[prevHash.String()] = &dl.entries[i]
		consistencyHashes[i] = prevHash.Clone()
	}
	dl.hash = prevHash
	dl.consistencyHashes = consistencyHashes
}

func (dl *documentLog) ConsistencyHashes() []model.Hash {
	return dl.consistencyHashes
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
func (dl documentLog) resolveAdvertedHashes() {
	// TODO: When to quite the loop?
	queue := dl.protocol.ReceivedConsistencyHashes()
	for {
		peerHash := queue.Get()
		log.Log().Infof("Got consistency hash (ours: %s, received: %s)", dl.hash.String(), peerHash.Hash.String())
		if dl.consistencyHashIndex[peerHash.Hash.String()] == nil {
			log.Log().Infof("Received unknown consistency hash, will query for document hash list (peer=%s,hash=%s)", peerHash.Peer, peerHash.Hash)
			// TODO: Don't have multiple parallel queries for the same peer / hash
			if err := dl.protocol.QueryHashList(peerHash.Peer); err != nil {
				log.Log().Errorf("Could query peer for hash list (peer=%s)", peerHash.Peer, err)
			}
		} else {
			log.Log().Info("Received known consistency hash, no action is required")
		}
	}
}

func (dl *documentLog) advertHash() {
	for {
		<-dl.advertHashTimer.C
		if !dl.hash.Empty() {
			log.Log().Infof("Adverting last hash (%s)", dl.hash)
			dl.protocol.AdvertConsistencyHash(dl.hash)
		}
	}
}
