package p2p

import (
	"errors"
	"github.com/nuts-foundation/nuts-network/pkg/model"
)

const ProtocolVersion = 1

var ErrMissingProtocolVersion = errors.New("missing protocol version")

var ErrUnsupportedProtocolVersion = errors.New("unsupported protocol version")

type P2PNetwork interface {
	// Start starts the P2P network on the local node.
	Start(config P2PNetworkConfig) error
	// Stop stops the P2P network on the local node.
	Stop() error
	// AddRemoteNode adds a remote node to the local node's view of the network, so it can become one of our peers.
	AddRemoteNode(node model.NodeInfo)
}

type P2PNetworkConfig struct {
	ListenAddress string
}
