package doclog

import (
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/p2p"
	"sort"
	"time"
)

const AdvertHashInterval = 2 * time.Second

func NewDocumentLog(p2pNetwork p2p.P2PNetwork) DocumentLog {
	return &documentLog{
		p2pNetwork:        p2pNetwork,
		hash:              model.EmptyHash(),
		consistencyHashes: make([]model.Hash, 0),
	}
}

type entry struct {
	// TODO: Cache consistency hash
	// hash contains the (cached) document hash
	hash model.Hash
	doc  *model.Document
}

type documentLog struct {
	p2pNetwork           p2p.P2PNetwork
	entries              []entry
	consistencyHashes    []model.Hash
	consistencyHashIndex map[string]*entry
	hash                 model.Hash
	advertHashTimer      *time.Ticker
}

func (dl *documentLog) Add(document *model.Document) {
	dl.entries = append(dl.entries, entry{
		hash: document.Hash(),
		doc:  document,
	})
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
	dl.p2pNetwork.SetHashSource(dl)
	dl.advertHashTimer = time.NewTicker(AdvertHashInterval)
	go dl.advertHash()
	go dl.resolveAdvertedHashes()
}

// TODO: Comment
// resolveAdvertedHashes reads
func (dl documentLog) resolveAdvertedHashes() {
	// TODO: When to quite the loop?
	queue := dl.p2pNetwork.ReceivedConsistencyHashes()
	for {
		peerHash := queue.Get()
		log.Log().Infof("Got consistency hash (ours: %s, received: %s)", dl.hash.String(), peerHash.Hash.String())
		if dl.consistencyHashIndex[peerHash.Hash.String()] == nil {
			log.Log().Infof("Received unknown consistency hash, will query for document hash list (peer=%s,hash=%s)", peerHash.Peer, peerHash.Hash)
			// TODO: Don't have multiple parallel queries for the same peer / hash
			if err := dl.p2pNetwork.QueryHashList(peerHash.Peer); err != nil {
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
			dl.p2pNetwork.AdvertConsistencyHash(dl.hash)
		}
	}
}
