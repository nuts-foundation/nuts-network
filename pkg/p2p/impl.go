package p2p

import (
	"context"
	"errors"
	"fmt"
	"github.com/nuts-foundation/nuts-go-core"
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/network"
	"github.com/nuts-foundation/nuts-network/pkg/concurrency"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	errors2 "github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	grpcPeer "google.golang.org/grpc/peer"
	"net"
	"time"
)

type p2pNetwork struct {
	node       model.NodeInfo
	publicAddr string

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
	peerMutex        *concurrency.SaferRWMutex
	receivedMessages messageQueue
}

func (n p2pNetwork) Diagnostics() []core.DiagnosticResult {
	peers := n.Peers()
	return []core.DiagnosticResult{
		NumberOfDiagnosticsResult{NumberOfPeers: len(peers)},
		PeersDiagnosticsResult{Peers: peers},
	}
}

func (n *p2pNetwork) Peers() []Peer {
	var result []Peer
	n.peerMutex.ReadLock(func() {
		for _, peer := range n.peers {
			result = append(result, Peer{
				NodeID:  peer.nodeId,
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
	connecting bool
}

func NewP2PNetwork() P2PNetwork {
	return &p2pNetwork{
		peers:                make(map[model.PeerID]*peer, 0),
		peersByAddr:          make(map[string]*peer, 0),
		remoteNodes:          make(map[model.NodeID]*remoteNode, 0),
		remoteNodeAddChannel: make(chan model.NodeInfo, 100), // TODO: Does this number make sense?
		peerMutex:            &concurrency.SaferRWMutex{},
		receivedMessages:     messageQueue{c: make(chan PeerMessage, 100)}, // TODO: Does this number make sense?
	}
}

type messageQueue struct {
	c chan PeerMessage
}

func (m messageQueue) Get() PeerMessage {
	return <-m.c
}

func (n *p2pNetwork) Start(config P2PNetworkConfig) error {
	log.Log().Infof("Starting gRPC server (ID: %s) on %s", config.NodeID, config.ListenAddress)
	n.publicAddr = config.PublicAddress
	var err error
	n.node.ID = config.NodeID
	n.listener, err = net.Listen("tcp", config.ListenAddress)
	if err != nil {
		return err
	}
	n.grpcServer = grpc.NewServer()
	network.RegisterNetworkServer(n.grpcServer, n)
	// TODO enable TLS
	// Start server
	go func() {
		err = n.grpcServer.Serve(n.listener)
		if err != nil {
			log.Log().Warn("grpcServer.Serve ended with err: ", err) // TODO: When does this happen?
		}
	}()
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

func (n p2pNetwork) AddRemoteNode(nodeInfo model.NodeInfo) {
	if n.shouldConnectTo(nodeInfo) {
		n.remoteNodeAddChannel <- nodeInfo
		log.Log().Infof("Added remote node to connect to: %s", nodeInfo)
	}
}

func (n *p2pNetwork) connectToNode(nodeInfo *model.NodeInfo) error {
	log.Log().Infof("Connecting to node: %s", nodeInfo)
	// TODO: Is this the right context?
	cxt := metadata.NewOutgoingContext(context.Background(), constructMetadata(n.node.ID))
	conn, err := grpc.DialContext(cxt, nodeInfo.Address, grpc.WithInsecure()) // TODO: Add TLS
	if err != nil {
		return err
	}
	// TODO: What if two node propagate the same ID? Maybe we shouldn't index our peers based on NodeID?
	client := network.NewNetworkClient(conn)
	gate, err := client.Connect(cxt)
	if err != nil {
		log.Log().Errorf("Failed to set up stream (node=%s): %v", nodeInfo, err)
		_ = conn.Close()
		return err
	}
	peer := &peer{
		id:     model.GetPeerID(nodeInfo.Address),
		nodeId: nodeInfo.ID,
		conn:   conn,
		client: client,
		gate:   gate,
		addr:   nodeInfo.Address,
	}
	n.sendAndReceiveForPeer(peer)
	return nil
}

func (n *p2pNetwork) sendAndReceiveForPeer(peer *peer) {
	go peer.startSending()
	n.addPeer(peer)
	// TODO: Check NodeID sent by peer
	peer.startReceiving(peer, n.receivedMessages)
	peer.close()
	// When we reach this line, startReceiving has exited which means the connection has been closed.
	n.removePeer(peer)
}

// connectToRemoteNodes reads from remoteNodeAddChannel to add remote nodes and connect to them
func (n *p2pNetwork) connectToRemoteNodes() {
	remoteNodesMutex := concurrency.SaferRWMutex{}
	go func() {
		for nodeInfo := range n.remoteNodeAddChannel {
			remoteNodesMutex.WriteLock(func() {
				if n.remoteNodes[nodeInfo.ID] == nil {
					n.remoteNodes[nodeInfo.ID] = &remoteNode{
						NodeInfo:   nodeInfo,
						connecting: false,
					}
					log.Log().Infof("Added remote node: %s", nodeInfo)
				}
			})
		}
	}()

	// TODO: We should probably do this per node
	// TODO: We need a backoff strategy
	ticker := time.NewTicker(1 * time.Second)
	for {
		// TODO: Exit strategy
		<-ticker.C
		remoteNodesMutex.ReadLock(func() {
			for _, remoteNode := range n.remoteNodes {
				if !remoteNode.connecting && n.shouldConnectTo(remoteNode.NodeInfo) {
					remoteNode.connecting = true
					go func() {
						if err := n.connectToNode(&remoteNode.NodeInfo); err != nil {
							log.Log().Warnf("Couldn't connect to node (node=%s): %v", remoteNode.NodeInfo, err)
						} else {
							log.Log().Infof("Connected to node: %s", remoteNode.NodeInfo)
						}
						remoteNode.connecting = false
					}()
				}
			}
		})
	}
}

// shouldConnectTo checks whether we should connect to the given node.
func (n p2pNetwork) shouldConnectTo(nodeInfo model.NodeInfo) bool {
	if normalizeAddress(nodeInfo.Address) == normalizeAddress(n.publicAddr) {
		// We're not going to connect to our own node
		log.Log().Debug("Not connecting since it's localhost")
		return false
	}
	var result = true
	n.peerMutex.ReadLock(func() {
		if nodeInfo.ID != "" {
			for _, peer := range n.peers {
				if peer.nodeId == nodeInfo.ID {
					// We're not going to connect to a node we're already connected to
					log.Log().Debugf("Not connecting since we're already connected (NodeID=%s)", nodeInfo.ID)
					result = false
					return
				}
			}
		}
		if n.peersByAddr[normalizeAddress(nodeInfo.Address)] != nil {
			// We're not going to connect to a node we're already connected to
			log.Log().Debugf("Not connecting since we're already connected (address=%s)", nodeInfo.Address)
			result = false
		}
	})
	return result
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
	nodeId, err := nodeIdFromMetadata(md)
	if err != nil {
		return err
	}
	// We received our peer's NodeID, now send our own.
	if err := stream.SendHeader(constructMetadata(n.node.ID)); err != nil {
		return errors2.Wrap(err, "unable to send headers")
	}
	peer := &peer{
		// TODO
		id:     model.GetPeerID(peerCtx.Addr.String()),
		nodeId: nodeId,
		gate:   stream,
		addr:   peerCtx.Addr.String(),
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


