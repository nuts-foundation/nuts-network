package p2p

import (
	"context"
	"fmt"
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/network"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net"
	"time"
)

type p2pNetwork struct {
	node model.NodeInfo
	// TODO: What if no-one is actually listening to this queue? Maybe we should create it when someone asks for it (lazy initialization)?
	receivedConsistencyHashes *AdvertedHashQueue
	receivedDocumentHashes    *AdvertedHashQueue
	/* gRPC server */
	grpcServer *grpc.Server
	listener   net.Listener
	/* gRPC client */
	// remoteNodes is the list of nodes (which we know of) that are part of the network. This doesn't mean they're
	// online per se or that we're even connected to them, just that we could try to do so.
	remoteNodes map[model.NodeID]*model.NodeInfo
	// remoteNodeAddChannel is used to communicate remote nodes we'd like to add (to remoteNodes)
	remoteNodeAddChannel chan model.NodeInfo // TODO: Do we actually need this channel or can we just spawn a goroutine instead?
	// peers is the list of nodes we're actually connected to.
	peers      map[model.PeerID]*peer
	hashSource HashSource
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

func (n *p2pNetwork) SetHashSource(source HashSource) {
	if n.hashSource != nil {
		panic("Hash source has already been set!")
	}
	n.hashSource = source
}

func (n p2pNetwork) ReceivedConsistencyHashes() PeerHashQueue {
	return n.receivedConsistencyHashes
}

func (n p2pNetwork) ReceivedDocumentHashes() PeerHashQueue {
	return n.receivedDocumentHashes
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

func (p peer) String() string {
	return fmt.Sprintf("%s(%s)", p.nodeId, p.addr)
}

func NewP2PNetwork() P2PNetwork {
	return &p2pNetwork{
		peers:                make(map[model.PeerID]*peer, 0),
		remoteNodes:          make(map[model.NodeID]*model.NodeInfo, 0),
		remoteNodeAddChannel: make(chan model.NodeInfo, 100), // TODO: Does this number make sense?
		receivedConsistencyHashes: &AdvertedHashQueue{
			c: make(chan PeerHash, 100), // TODO: Does this number make sense?
		},
		receivedDocumentHashes: &AdvertedHashQueue{
			c: make(chan PeerHash, 1000), // TODO: Does this number make sense?
		},
	}
}

func (n *p2pNetwork) Start(config P2PNetworkConfig) error {
	log.Log().Infof("Starting gRPC server (ID: %s) on %s", config.NodeID, config.ListenAddress)
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

func (n *p2pNetwork) AddRemoteNode(nodeInfo model.NodeInfo) {
	n.remoteNodeAddChannel <- nodeInfo
	log.Log().Infof("Added remote node to connect to: %s", nodeInfo)
}

func (n p2pNetwork) AdvertConsistencyHash(hash model.Hash) {
	// TODO: Synchronization on peers?
	msg := createMessage()
	msg.AdvertHash = &network.AdvertHash{Hash: hash}
	for _, peer := range n.peers {
		if err := peer.gate.Send(&msg); err != nil {
			log.Log().Errorf("Unable to send message (peer=%s): %v", peer, err)
			// TODO: What now?
		}
	}
}

func (n p2pNetwork) QueryHashList(peerId model.PeerID) error {
	peer := n.peers[peerId]
	if peer == nil {
		return fmt.Errorf("unknown peer: %s", peerId)
	}
	msg := createMessage()
	msg.HashListQuery = &network.HashListQuery{}
	return peer.gate.Send(&msg)
}

func createMessage() network.NetworkMessage {
	return network.NetworkMessage{
		Header: &network.Header{
			Version: ProtocolVersion,
		},
	}
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
	go func() {
		// We can safely ignore the error since all handling (cleaning up after an error) is done by receiveFromPeer
		_ = n.receiveFromPeer(peer, gate)
	}()
	return nil
}

// connectToRemoteNodes reads from remoteNodeAddChannel to add remote nodes and connect to them
func (n *p2pNetwork) connectToRemoteNodes() {
	go func() {
		for nodeInfo := range n.remoteNodeAddChannel {
			if n.remoteNodes[nodeInfo.ID] == nil {
				n.remoteNodes[nodeInfo.ID] = &nodeInfo
				log.Log().Infof("Added remote node: %s", nodeInfo)
			}
		}
	}()

	// TODO: We should probably do this per node
	// TODO: We need a backoff strategy
	ticker := time.NewTicker(1 * time.Second)
	for {
		// TODO: Exit strategy
		<-ticker.C
		for _, nodeInfo := range n.remoteNodes {
			if !n.isConnectedTo(*nodeInfo) {
				if err := n.connectToNode(nodeInfo); err != nil {
					log.Log().Warnf("Couldn't connect to node (node=%s): %v", nodeInfo, err)
				}
			}
		}
	}
}

// isConnectedTo checks whether we're currently connected to the given node.
func (n p2pNetwork) isConnectedTo(nodeInfo model.NodeInfo) bool {
	for _, peer := range n.peers {
		if peer.nodeId == nodeInfo.ID {
			return true
		}
	}
	return false
}

func (n p2pNetwork) isRunning() bool {
	return n.grpcServer != nil
}

//func (client networkClient) getLocalHostnames() ([]string, error) {
//	var result []string
//	ifaces, err := net.Interfaces()
//	if err != nil {
//		return nil, err
//	}
//	// handle err
//	for _, i := range ifaces {
//		addrs, err := i.Addrs()
//		if err != nil {
//			return nil, err
//		}
//		// handle err
//		for _, addr := range addrs {
//			var ip net.IP
//			switch v := addr.(type) {
//			case *net.IPNet:
//				ip = v.IP
//			case *net.IPAddr:
//				ip = v.IP
//			}
//			if ip != nil {
//				result = append(result, ip.String())
//			} else {
//				// TODO
//			}
//		}
//	}
//	return result, nil
//}
