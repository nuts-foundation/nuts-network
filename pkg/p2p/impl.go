package p2p

import (
	"context"
	"errors"
	"fmt"
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/network"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	errors2 "github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	grpcPeer "google.golang.org/grpc/peer"
	"io"
	"net"
	"strings"
	"time"
)

type p2pNetwork struct {
	node model.NodeInfo

	/* gRPC server */
	grpcServer *grpc.Server
	listener   net.Listener
	/* gRPC client */
	// remoteNodes is the list of nodes (which we know of) that are part of the network. This doesn't mean they're
	// online per se or that we're even connected to them, just that we could try to do so.
	remoteNodes map[model.NodeID]*remoteNode
	// remoteNodeAddChannel is used to communicate remote nodes we'd like to add (to remoteNodes)
	remoteNodeAddChannel chan model.NodeInfo // TODO: Do we actually need this channel or can we just spawn a goroutine instead?
	// peers is the list of nodes we're actually connected to.
	peers            map[model.PeerID]*peer
	peersByAddr      map[string]*peer
	receivedMessages messageQueue
	publicAddr       string
}

func (n p2pNetwork) Broadcast(message *network.NetworkMessage) {
	for _, peer := range n.peers {
		if peer.gate.Send(message) != nil {
			log.Log().Errorf("Unable to broadcast message to peer (peer=%s)", peer.id)
		}
	}
}

func (n p2pNetwork) ReceivedMessages() MessageQueue {
	return n.receivedMessages
}

func (n p2pNetwork) Send(peerId model.PeerID, message *network.NetworkMessage) error {
	peer := n.peers[peerId]
	if peer == nil {
		return fmt.Errorf("unknown peer: %s", peerId)
	}
	return peer.gate.Send(message)
}

// Peer represents a connected peer
type peer struct {
	id     model.PeerID
	nodeId model.NodeID
	client network.NetworkClient
	gate   messageGate
	// conn is only filled for peers where we're the connecting party
	conn *grpc.ClientConn
	addr string
}

type remoteNode struct {
	model.NodeInfo
	connecting bool
}

func (p peer) String() string {
	return fmt.Sprintf("%s(%s)", p.nodeId, p.addr)
}

func NewP2PNetwork() P2PNetwork {
	return &p2pNetwork{
		peers:                make(map[model.PeerID]*peer, 0),
		peersByAddr:          make(map[string]*peer, 0),
		remoteNodes:          make(map[model.NodeID]*remoteNode, 0),
		remoteNodeAddChannel: make(chan model.NodeInfo, 100),               // TODO: Does this number make sense?
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
	for _, peer := range n.peers {
		if err := peer.conn.Close(); err != nil {
			log.Log().Errorf("Unable to close client connection (peer=%s): %v", peer, err)
		}
		n.removePeer(peer)
	}
	return nil
}

func (n p2pNetwork) AddRemoteNode(nodeInfo model.NodeInfo) {
	if !n.isConnectedTo(nodeInfo) {
		n.remoteNodeAddChannel <- nodeInfo
		log.Log().Infof("Added remote node to connect to: %s", nodeInfo)
	}
}

// receiveFromPeer reads messages from the peer until it disconnects or the network is stopped.
func (n p2pNetwork) receiveFromPeer(peer *peer, gate messageGate) {
	for n.isRunning() {
		msg, recvErr := gate.Recv()
		if recvErr != nil {
			if recvErr == io.EOF {
				log.Log().Infof("Peer closed connection: %s", peer)
			} else {
				log.Log().Errorf("Peer connection error (peer=%s): %v", peer, recvErr)
			}
			break
		}
		log.Log().Debugf("Received message from peer (%s): %s", peer, msg.String())
		n.receivedMessages.c <- PeerMessage{
			Peer:    peer.id,
			Message: msg,
		}
	}
	// TODO: Synchronization
	n.removePeer(peer)
}

func (n *p2pNetwork) connectToNode(nodeInfo *model.NodeInfo) error {
	// TODO: Synchronization
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
	n.addPeer(peer)
	// TODO: Check NodeID sent by peer
	go n.receiveFromPeer(peer, gate)
	return nil
}

// connectToRemoteNodes reads from remoteNodeAddChannel to add remote nodes and connect to them
func (n *p2pNetwork) connectToRemoteNodes() {
	go func() {
		for nodeInfo := range n.remoteNodeAddChannel {
			if n.remoteNodes[nodeInfo.ID] == nil {
				n.remoteNodes[nodeInfo.ID] = &remoteNode{
					NodeInfo:   nodeInfo,
					connecting: false,
				}
				log.Log().Infof("Added remote node: %s", nodeInfo)
			}
		}
	}()

	// TODO: We should probably do this per node
	// TODO: We need a backoff strategy
	ticker := time.NewTicker(1 * time.Second)
	for {
		// TODO: Exit strategy
		// TODO: Synchronize
		<-ticker.C
		for _, remoteNode := range n.remoteNodes {
			if !remoteNode.connecting && !n.isConnectedTo(remoteNode.NodeInfo) {
				remoteNode.connecting = true
				if err := n.connectToNode(&remoteNode.NodeInfo); err != nil {
					log.Log().Warnf("Couldn't connect to node (node=%s): %v", remoteNode.NodeInfo, err)
					remoteNode.connecting = false
				}
			}
		}
	}
}

// isConnectedTo checks whether we're currently connected to the given node.
func (n p2pNetwork) isConnectedTo(nodeInfo model.NodeInfo) bool {
	log.Log().Infof("Might connect to remote node %s, own address: %s, current peers: %v", nodeInfo, n.publicAddr, n.peersByAddr)
	if nodeInfo.Address == n.publicAddr {
		// We're not going to connect to our own node
		log.Log().Info("Not connecting since it's localhost")
		return true
	}
	// Check connected to node ID?
	if nodeInfo.ID != "" {
		for _, peer := range n.peers {
			if peer.nodeId == nodeInfo.ID {
				// We're not going to connect to a node we're already connected to
				log.Log().Infof("Not connecting since we're already connected (nodeID=%s)", nodeInfo.ID)
				return true
			}
		}
	}
	if n.isConnectedToAddress(nodeInfo.Address) {
		// We're not going to connect to a node we're already connected to
		log.Log().Infof("Not connecting since we're already connected (address=%s)", nodeInfo.Address)
		return true
	}
	return false
}

func (n p2pNetwork) isRunning() bool {
	return n.grpcServer != nil
}

const NodeIDHeader = "nodeId"

type messageGate interface {
	Send(message *network.NetworkMessage) error
	Recv() (*network.NetworkMessage, error)
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
	n.addPeer(peer)
	n.receiveFromPeer(peer, stream)
	return nil
}

func nodeIdFromMetadata(md metadata.MD) (model.NodeID, error) {
	vals := md.Get(NodeIDHeader)
	if len(vals) == 0 {
		return "", fmt.Errorf("peer didn't send %s header", NodeIDHeader)
	} else if len(vals) > 1 {
		return "", fmt.Errorf("peer sent multiple values for %s header", NodeIDHeader)
	}
	return model.NodeID(strings.TrimSpace(vals[0])), nil
}

func constructMetadata(nodeId model.NodeID) metadata.MD {
	return metadata.New(map[string]string{NodeIDHeader: string(nodeId)})
}

func (n *p2pNetwork) addPeer(peer *peer) {
	n.peers[peer.id] = peer
	n.peersByAddr[normalizeAddress(peer.addr)] = peer
}

func (n p2pNetwork) isConnectedToAddress(addr string) bool {
	return n.peersByAddr[normalizeAddress(addr)] != nil
}

func normalizeAddress(addr string) string {
	var normalizedAddr string
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		normalizedAddr = addr
	} else {
		if host == "localhost" {
			host = "127.0.0.1"
			normalizedAddr = net.JoinHostPort(host, port)
		} else {
			normalizedAddr = addr
		}
	}
	return normalizedAddr
}

func (n *p2pNetwork) removePeer(peer *peer) {
	delete(n.peers, peer.id)
	delete(n.peersByAddr, peer.addr)
}
