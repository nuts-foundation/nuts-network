/*
 * Copyright (C) 2020. Nuts community
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 *
 */

package p2p

import (
	"crypto/tls"
	"fmt"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	"github.com/nuts-foundation/nuts-network/network"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/stats"
)

// P2PNetwork defines the API for the P2P layer, used to connect to peers and exchange messages.
type P2PNetwork interface {
	stats.StatsProvider
	// Configure configures the P2PNetwork. Must be called before Start().
	Configure(config P2PNetworkConfig) error
	// Start starts the P2P network on the local node.
	Start() error
	// Stop stops the P2P network on the local node.
	Stop() error
	// AddRemoteNode adds a remote node to the local node's view of the network, so it can become one of our peers.
	AddRemoteNode(node model.NodeInfo)
	// ReceivedMessages returns a queue containing all messages received from our peers. It must be drained, because when its buffer is full the producer (P2PNetwork) is blocked.
	ReceivedMessages() MessageQueue
	// Send sends a message to a specific peer.
	Send(peer model.PeerID, message *network.NetworkMessage) error
	// Broadcast sends a message to all peers.
	Broadcast(message *network.NetworkMessage)
	// Peers returns our peers (remote nodes we're currently connected to).
	Peers() []Peer
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
	NodeID         model.NodeID
	PublicAddress  string
	ListenAddress  string
	BootstrapNodes []string
	ClientCert     tls.Certificate
	ServerCert     tls.Certificate
	TrustStore     cert.TrustStore
}
