package p2p

import (
	"crypto/tls"
	"fmt"
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-network/network"
	"github.com/nuts-foundation/nuts-network/pkg/model"
)

type P2PNetwork interface {
	// Start starts the P2P network on the local node.
	Start(config P2PNetworkConfig) error
	// Stop stops the P2P network on the local node.
	Stop() error
	// AddRemoteNode adds a remote node to the local node's view of the network, so it can become one of our peers.
	AddRemoteNode(node model.NodeInfo)
	ReceivedMessages() MessageQueue
	Send(peer model.PeerID, message *network.NetworkMessage) error
	Broadcast(message *network.NetworkMessage)
	// Peers returns the peers we're currently connected to
	Peers() []Peer
	Diagnostics() []core.DiagnosticResult
}

type MessageQueue interface {
	Get() PeerMessage
}

type Peer struct {
	NodeID  model.NodeID
	PeerID  model.PeerID
	Address string
}

func (p Peer) String() string {
	if p.NodeID != "" {
		return fmt.Sprintf("(ID=%s,NodeID=%s,Addr=%s)", p.PeerID, p.NodeID, p.Address)
	}
	return p.Address
}

type PeerMessage struct {
	Peer    model.PeerID
	Message *network.NetworkMessage
}

type P2PNetworkConfig struct {
	NodeID        model.NodeID
	PublicAddress string
	ListenAddress string
	ClientCert    tls.Certificate
	ServerCert    tls.Certificate
}
