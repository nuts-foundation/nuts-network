package p2p

import (
	"errors"
	"github.com/nuts-foundation/nuts-network/pkg/model"
)

const ProtocolVersion = 1

const NodeIDHeader = "nodeId"

var ErrMissingProtocolVersion = errors.New("missing protocol version")

var ErrUnsupportedProtocolVersion = errors.New("unsupported protocol version")

type P2PNetwork interface {
	// Start starts the P2P network on the local node.
	Start(config P2PNetworkConfig) error
	// Stop stops the P2P network on the local node.
	Stop() error
	// AddRemoteNode adds a remote node to the local node's view of the network, so it can become one of our peers.
	AddRemoteNode(node model.NodeInfo)
	// Sets source for hashes (documents, document hashes and consistency hashes), so we can provide our peers with it when queried.
	// TODO: This feels like poor man's dependency injection... We need design pattern here.
	SetHashSource(source HashSource)
	AdvertConsistencyHash(hash model.Hash)
	QueryHashList(peer model.PeerID) error
	ReceivedConsistencyHashes() PeerHashQueue
	ReceivedDocumentHashes() PeerHashQueue
}

type HashSource interface {
	ConsistencyHashes() []model.Hash
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

type P2PNetworkConfig struct {
	NodeID        model.NodeID
	ListenAddress string
}
