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
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/network"
	"github.com/nuts-foundation/nuts-network/pkg/concurrency"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/stats"
	errors2 "github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	grpcPeer "google.golang.org/grpc/peer"
	"net"
	"strings"
	"sync"
	"time"
)

type Dialer func(ctx context.Context, target string, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error)

const addRemoteNodeChannelSize = 100

type p2pNetwork struct {
	config P2PNetworkConfig

	grpcServer *grpc.Server
	listener   net.Listener

	// remoteNodes is the list of nodes (which we know of) that are part of the network. This doesn't mean they're
	// online per se or that we're even connected to them, just that we could try to do so.
	remoteNodes map[model.NodeID]*remoteNode
	// remoteNodeAddChannel is used to communicate remote nodes we'd like to add (to remoteNodes)
	remoteNodeAddChannel chan model.NodeInfo // TODO: Do we actually need this channel or can we just spawn a goroutine instead?
	// peers is the list of nodes we're actually connected to. Access MUST be wrapped in locking using peerReadLock and peerWriteLock!
	peers map[model.PeerID]*peer
	// peersByAddr Access MUST be wrapped in locking using peerReadLock and peerWriteLock
	peersByAddr      map[string]*peer
	peerMutex        concurrency.SaferRWMutex
	receivedMessages messageQueue
	peerDialer       Dialer
	configured       bool
}

func (n p2pNetwork) Configured() bool {
	return n.configured
}

func (n p2pNetwork) Statistics() []stats.Statistic {
	peers := n.Peers()
	return []stats.Statistic{
		NumberOfPeersStatistic{NumberOfPeers: len(peers)},
		PeersStatistic{Peers: peers},
	}
}

func (n *p2pNetwork) Peers() []Peer {
	var result []Peer
	n.peerMutex.ReadLock(func() {
		for _, peer := range n.peers {
			result = append(result, Peer{
				NodeID:  peer.nodeID,
				PeerID:  peer.id,
				Address: peer.addr,
			})
		}
	})
	return result
}

func (n *p2pNetwork) Broadcast(message *network.NetworkMessage) {
	n.peerMutex.ReadLock(func() {
		for _, peer := range n.peers {
			peer.outMessages <- message
		}
	})
}

func (n p2pNetwork) ReceivedMessages() MessageQueue {
	return n.receivedMessages
}

func (n p2pNetwork) Send(peerId model.PeerID, message *network.NetworkMessage) error {
	// TODO: Can't we optimize this so that we don't need this lock? Maybe by (secretly) embedding a pointer to the peer in the peer ID?
	var peer *peer
	n.peerMutex.ReadLock(func() {
		peer = n.peers[peerId]
	})
	if peer == nil {
		return fmt.Errorf("unknown peer: %s", peerId)
	}
	peer.outMessages <- message
	return nil
}

type remoteNode struct {
	model.NodeInfo
	backoff Backoff
	Dialer
}

func (r *remoteNode) connect(nodeId model.NodeID, config *tls.Config) (*peer, error) {
	log.Log().Infof("Connecting to node: %s", r.NodeInfo)
	cxt := metadata.NewOutgoingContext(context.Background(), constructMetadata(nodeId))
	dialContext, _ := context.WithTimeout(cxt, 5*time.Second)
	conn, err := r.Dialer(dialContext, r.NodeInfo.Address,
		grpc.WithBlock(), // Dial should block until connection succeeded (or time-out expired)
		grpc.WithTransportCredentials(credentials.NewTLS(config)), // TLS authentication
		grpc.WithReturnConnectionError()) // This option causes underlying errors to be returned when connections fail, rather than just "context deadline exceeded"
	if err != nil {
		return nil, errors2.Wrap(err, "unable to connect")
	}
	// TODO: What if two node propagate the same ID? Maybe we shouldn't index our peers based on NodeID?
	client := network.NewNetworkClient(conn)
	gate, err := client.Connect(cxt)
	if err != nil {
		log.Log().Errorf("Failed to set up stream (node=%s): %v", r.NodeInfo, err)
		_ = conn.Close()
		return nil, err
	}
	if serverHeader, err := gate.Header(); err != nil {
		log.Log().Errorf("Error receiving headers from server (node=%s): %v", r.NodeInfo, err)
		_ = conn.Close()
		return nil, err
	} else {
		if serverNodeID, err := nodeIDFromMetadata(serverHeader); err != nil {
			log.Log().Errorf("Error parsing NodeID header from server (node=%s): %v", r.NodeInfo, err)
			_ = conn.Close()
			return nil, err
		} else {
			if !r.NodeInfo.ID.Empty() && r.NodeInfo.ID != serverNodeID {
				// TODO: What to do here?
				log.Log().Warnf("Server sent different NodeID than expected (expected=%s,actual=%s)", r.NodeInfo.ID, serverNodeID)
			} else {
				r.NodeInfo.ID = serverNodeID
			}
		}
	}

	return &peer{
		id:         model.GetPeerID(r.NodeInfo.Address),
		nodeID:     r.NodeInfo.ID,
		conn:       conn,
		client:     client,
		gate:       gate,
		addr:       r.NodeInfo.Address,
		closeMutex: &sync.Mutex{},
	}, nil
}

func NewP2PNetwork() P2PNetwork {
	return &p2pNetwork{
		peers:                make(map[model.PeerID]*peer, 0),
		peersByAddr:          make(map[string]*peer, 0),
		remoteNodes:          make(map[model.NodeID]*remoteNode, 0),
		remoteNodeAddChannel: make(chan model.NodeInfo, addRemoteNodeChannelSize), // TODO: Does this number make sense?
		peerMutex:            concurrency.NewSaferRWMutex("p2p-peers"),
		receivedMessages:     messageQueue{c: make(chan PeerMessage, 100)}, // TODO: Does this number make sense?
		peerDialer:           grpc.DialContext,
	}
}

func NewP2PNetworkWithOptions(listener net.Listener, dialer Dialer) P2PNetwork {
	result := NewP2PNetwork().(*p2pNetwork)
	result.listener = listener
	result.peerDialer = dialer
	return result
}

type messageQueue struct {
	c chan PeerMessage
}

func (m messageQueue) Get() PeerMessage {
	return <-m.c
}

func (n *p2pNetwork) Configure(config P2PNetworkConfig) error {
	if config.NodeID == "" {
		return errors.New("NodeID is empty")
	}
	if config.TrustStore == nil {
		return errors.New("TrustStore is nil")
	}
	n.config = config
	n.configured = true
	for _, bootstrapNode := range n.config.BootstrapNodes {
		n.AddRemoteNode(model.NodeInfo{Address: bootstrapNode})
	}
	return nil
}

func (n *p2pNetwork) Start() error {
	log.Log().Infof("Starting gRPC P2P node (ID: %s)", n.config.NodeID)
	if n.config.ListenAddress != "" {
		log.Log().Infof("Starting gRPC server on %s", n.config.ListenAddress)
		var err error
		// We allow test code to set the listener to allow for in-memory (bufnet) channels
		var serverOpts = make([]grpc.ServerOption, 0)
		if n.listener == nil {
			n.listener, err = net.Listen("tcp", n.config.ListenAddress)
			if err != nil {
				return err
			}
			if n.config.ServerCert.PrivateKey == nil {
				log.Log().Info("TLS is disabled on gRPC server side! Make sure SSL/TLS offloading is properly configured.")
			} else {
				_, rootPool := n.config.TrustStore.Roots()
				serverOpts = append(serverOpts, grpc.Creds(credentials.NewTLS(&tls.Config{
					Certificates: []tls.Certificate{n.config.ServerCert},
					ClientAuth:   tls.RequireAndVerifyClientCert,
					ClientCAs:    rootPool,
				})))
			}
		}
		n.grpcServer = grpc.NewServer(serverOpts...)
		network.RegisterNetworkServer(n.grpcServer, n)
		go func() {
			err = n.grpcServer.Serve(n.listener)
			if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
				log.Log().Errorf("gRPC server errored: %v", err)
				_ = n.Stop()
			}
		}()
	}
	// Start client
	go n.connectToRemoteNodes()
	return nil
}

func (n *p2pNetwork) Stop() error {
	// Stop server
	if n.grpcServer != nil {
		n.grpcServer.Stop()
		n.grpcServer = nil
	}
	if n.listener != nil {
		if err := n.listener.Close(); err != nil {
			log.Log().Warn("Error while closing server listener: ", err)
		}
		n.listener = nil
	}
	close(n.remoteNodeAddChannel)
	// Stop client
	n.peerMutex.ReadLock(func() {
		for _, peer := range n.peers {
			peer.close()
		}
	})
	return nil
}

func (n p2pNetwork) AddRemoteNode(nodeInfo model.NodeInfo) bool {
	if n.shouldConnectTo(nodeInfo) && len(n.remoteNodeAddChannel) < addRemoteNodeChannelSize {
		n.remoteNodeAddChannel <- nodeInfo
		return true
	}
	return false
}

func (n *p2pNetwork) sendAndReceiveForPeer(peer *peer) {
	peer.outMessages = make(chan *network.NetworkMessage, 10) // TODO: Does this number make sense? Should also be configurable?
	go peer.sendMessages()
	n.addPeer(peer)
	// TODO: Check NodeID sent by peer
	receiveMessages(peer.gate, peer.id, n.receivedMessages)
	peer.close()
	// When we reach this line, receiveMessages has exited which means the connection has been closed.
	n.removePeer(peer)
}

// connectToRemoteNodes reads from remoteNodeAddChannel to add remote nodes and connect to them
func (n *p2pNetwork) connectToRemoteNodes() {
	for nodeInfo := range n.remoteNodeAddChannel {
		if n.remoteNodes[nodeInfo.ID] == nil {
			remoteNode := &remoteNode{
				NodeInfo: nodeInfo,
				backoff:  defaultBackoff(),
				Dialer:   n.peerDialer,
			}
			n.remoteNodes[nodeInfo.ID] = remoteNode
			log.Log().Infof("Added remote node: %s", nodeInfo)
			go func() {
				for {
					if n.shouldConnectTo(remoteNode.NodeInfo) {
						tlsConfig := tls.Config{
							Certificates: []tls.Certificate{n.config.ClientCert},
						}
						if peer, err := remoteNode.connect(n.config.NodeID, &tlsConfig); err != nil {
							waitPeriod := remoteNode.backoff.Backoff()
							log.Log().Warnf("Couldn't connect to node, reconnecting in %d seconds (node=%s,err=%v)", int(waitPeriod.Seconds()), remoteNode.NodeInfo, err)
							time.Sleep(waitPeriod)
						} else {
							n.sendAndReceiveForPeer(peer)
							remoteNode.backoff.Reset()
							log.Log().Infof("Connected to node: %s", remoteNode.NodeInfo)
						}
					}
					time.Sleep(5 * time.Second)
				}
			}()
		}
	}
}

// shouldConnectTo checks whether we should connect to the given node.
func (n p2pNetwork) shouldConnectTo(nodeInfo model.NodeInfo) bool {
	if normalizeAddress(nodeInfo.Address) == normalizeAddress(n.getLocalAddress()) {
		// We're not going to connect to our own node
		log.Log().Debug("Not connecting since it's localhost")
		return false
	}
	var result = true
	n.peerMutex.ReadLock(func() {
		if nodeInfo.ID != "" {
			for _, peer := range n.peers {
				if peer.nodeID == nodeInfo.ID {
					// We're not going to connect to a node we're already connected to
					log.Log().Tracef("Not connecting since we're already connected (NodeID=%s)", nodeInfo.ID)
					result = false
					return
				}
			}
		}
		if n.peersByAddr[normalizeAddress(nodeInfo.Address)] != nil {
			// We're not going to connect to a node we're already connected to
			log.Log().Tracef("Not connecting since we're already connected (address=%s)", nodeInfo.Address)
			result = false
		}
	})
	return result
}

func (n p2pNetwork) getLocalAddress() string {
	if n.config.PublicAddress != "" {
		return n.config.PublicAddress
	} else {
		if strings.HasPrefix(n.config.ListenAddress, ":") {
			// Interface's address not included in listening address (e.g. :5555), so prepend with localhost
			return "localhost" + n.config.ListenAddress
		} else {
			// Interface's address included in listening address (e.g. localhost:5555), so return as-is.
			return n.config.ListenAddress
		}
	}
}

func (n p2pNetwork) isRunning() bool {
	return n.grpcServer != nil
}

func (n p2pNetwork) Connect(stream network.Network_ConnectServer) error {
	peerCtx, _ := grpcPeer.FromContext(stream.Context())
	log.Log().Infof("New peer connected from %s", peerCtx.Addr)
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return errors.New("unable to get metadata")
	}
	nodeID, err := nodeIDFromMetadata(md)
	if err != nil {
		return err
	}
	// We received our peer's NodeID, now send our own.
	if err := stream.SendHeader(constructMetadata(n.config.NodeID)); err != nil {
		return errors2.Wrap(err, "unable to send headers")
	}
	remoteAddr := peerCtx.Addr.String()
	if remoteAddr == "bufconn" {
		// This is a shared-memory connection, in which case we should take the nodeID since all incoming shared-memory
		// connections have the same remote address.
		remoteAddr = nodeID.String()
	}
	peer := &peer{
		id:         model.GetPeerID(remoteAddr),
		nodeID:     nodeID,
		gate:       stream,
		addr:       remoteAddr,
		closeMutex: &sync.Mutex{},
	}
	n.sendAndReceiveForPeer(peer)
	return nil
}

func (n *p2pNetwork) addPeer(peer *peer) {
	n.peerMutex.WriteLock(func() {
		n.peers[peer.id] = peer
		n.peersByAddr[normalizeAddress(peer.addr)] = peer
	})
}

func (n *p2pNetwork) removePeer(peer *peer) {
	n.peerMutex.WriteLock(func() {
		peer = n.peers[peer.id]
		if peer == nil {
			return
		}

		delete(n.peers, peer.id)
		delete(n.peersByAddr, normalizeAddress(peer.addr))
	})
}
