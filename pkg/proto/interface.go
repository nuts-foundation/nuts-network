package proto

import (
	"errors"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/p2p"
)

const Version = 1

var ErrMissingProtocolVersion = errors.New("missing protocol version")

var ErrUnsupportedProtocolVersion = errors.New("unsupported protocol version")

type Protocol interface {
	// TODO: This feels like poor man's dependency injection... We need a design pattern here.
	Start(p2pNetwork p2p.P2PNetwork, source HashSource)
	Stop()

	ReceivedConsistencyHashes() PeerHashQueue
	ReceivedDocumentHashes() PeerHashQueue

	AdvertConsistencyHash(hash model.Hash)
	QueryHashList(peer model.PeerID) error
}

// PeerHashQueue is a queue which contains the hashes adverted by our peers. It's a FILO queue, since
// the hashes represent append-only data structures which means the last one is most recent.
type PeerHashQueue interface {
	// Get blocks until there's an PeerHash available and returns it.
	// TODO: Cancellation?
	Get() PeerHash
}

type PeerHash struct {
	Peer model.PeerID
	Hash model.Hash
}


type HashSource interface {
	ConsistencyHashes() []model.Hash
	ContainsDocument(hash model.Hash) bool
	Add(document *model.Document)
}

type AdvertedHashQueue struct {
	c chan PeerHash
}

func (q AdvertedHashQueue) Get() PeerHash {
	return <-q.c
}
func (q AdvertedHashQueue) put(hash PeerHash) {
	q.c <- hash
}