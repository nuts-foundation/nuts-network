package documents

import (
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/pkg/p2p"
	"time"
)

const AdvertLastHashInterval = 2 * time.Second

func NewDocumentLog(p2pNetwork p2p.P2PNetwork) DocumentLog {
	return &documentLog{p2pNetwork: p2pNetwork}
}

type documentLog struct {
	p2pNetwork          p2p.P2PNetwork
	lastHash            []byte
	advertLastHashTimer *time.Timer
}

func (dl *documentLog) Stop() {
	// TODO: Should check result of Stop()
	dl.advertLastHashTimer.Stop()
}

func (dl *documentLog) Start() {
	// TODO
	dl.lastHash = []byte("foobar")
	dl.advertLastHashTimer = time.NewTimer(AdvertLastHashInterval)
}

func (dl documentLog) advertLastHash() {
	for {
		<-dl.advertLastHashTimer.C
		log.Log().Info("Adverting last ash")
		dl.p2pNetwork.AdvertHash(dl.lastHash)
	}
}
