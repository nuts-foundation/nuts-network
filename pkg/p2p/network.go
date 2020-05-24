package p2p

import (
	"context"
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/network"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"google.golang.org/grpc"
	"io"
	"net"
	"time"
)

type p2pNetwork struct {
	ownAddr string
	/* gRPC server */
	grpcServer *grpc.Server
	listener   net.Listener
	/* gRPC client */
	// remoteNodes is the list of nodes (which we know of) that are part of the network. This doesn't mean they're
	// online per se or that we're even connected to them, just that we could try to do so.
	remoteNodes map[model.NodeID]*model.NodeInfo
	// remoteNodeAddChannel is used to communicate remote nodes we'd like to add (to remoteNodes)
	remoteNodeAddChannel chan model.NodeInfo
	// peers is the list of nodes we're actually connected to.
	peers map[model.NodeID]*peer
}

// Peer represents a connected peer
type peer struct {
	node   *model.NodeInfo
	conn   *grpc.ClientConn
	client network.NetworkClient
}

func NewP2PNetwork() P2PNetwork {
	return &p2pNetwork{
		peers:                make(map[model.NodeID]*peer, 0),
		remoteNodes:          make(map[model.NodeID]*model.NodeInfo, 0),
		remoteNodeAddChannel: make(chan model.NodeInfo, 100),
	}
}

func (n *p2pNetwork) Start(config P2PNetworkConfig) error {
	log.Log().Infof("Starting gRPC server on %s", config.ListenAddress)
	var err error
	n.listener, err = net.Listen("tcp", config.ListenAddress)
	if err != nil {
		return err
	}
	n.grpcServer = grpc.NewServer()
	network.RegisterNetworkServer(n.grpcServer, n)
	// TODO enable TLS
	go func() {
		err = n.grpcServer.Serve(n.listener)
		if err != nil {
			log.Log().Warn("grpcServer.Serve ended with err: ", err) // TODO: When does this happen?
		}
	}()
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
			log.Log().Errorf("Unable to close client connection (node: %s)", peer.node, err)
		}
		n.peers[peer.node.ID] = nil
	}
	return nil
}

func (n *p2pNetwork) AddRemoteNode(nodeInfo model.NodeInfo) {
	n.remoteNodeAddChannel <- nodeInfo
}

// connectToRemoteNodes reads from remoteNodeAddChannel to add remote nodes and connect to them
func (n *p2pNetwork) connectToRemoteNodes() {
	go func() {
		for nodeInfo := range n.remoteNodeAddChannel {
			if n.remoteNodes[nodeInfo.ID] != nil {
				n.remoteNodes[nodeInfo.ID] = &nodeInfo
				log.Log().Infof("Added remote node: %s", nodeInfo)
			}
		}
	}()

	// TODO: We should probably do this per node
	// TODO: We need a backoff strategy
	timer1 := time.NewTimer(1 * time.Second)
	for {
		// TODO: Exit strategy
		<-timer1.C
		for _, nodeInfo := range n.remoteNodes {
			if !n.isConnectedTo(*nodeInfo) {
				if err := n.connectToNode(nodeInfo); err != nil {
					log.Log().Warnf("Couldn't connect to node: %s", nodeInfo, err)
				}
			}
		}
	}
}

// isConnectedTo checks whether we're currently connected to the given node.
func (n p2pNetwork) isConnectedTo(nodeInfo model.NodeInfo) bool {
	return n.peers[nodeInfo.ID] != nil
}

func (n p2pNetwork) isRunning() bool {
	return n.grpcServer != nil
}

func (n *p2pNetwork) connectToNode(nodeInfo *model.NodeInfo) error {
	log.Log().Infof("Connecting to node: %s", nodeInfo)
	conn, err := grpc.Dial(nodeInfo.Address, grpc.WithInsecure()) // TODO: Add TLS
	if err != nil {
		return err
	}
	// TODO: What if two node propagate the same ID? Maybe we shouldn't index our peers based on NodeID?
	peer := &peer{conn: conn, client: network.NewNetworkClient(conn)}
	n.peers[nodeInfo.ID] = peer
	go n.receiveFromPeer(peer)
	return nil
}

// receiveFromPeer reads messages from the peer until it disconnects or the network is stopped.
func (n p2pNetwork) receiveFromPeer(peer *peer) {
	gateway, err := peer.client.Connect(context.Background())
	if err != nil {
		log.Log().Errorf("Failed to set up stream to peer: %s", peer.node, err)
		_ = peer.conn.Close()
		n.removePeer(peer)
		return
	}
	for n.isRunning() {
		msg, err := gateway.Recv()
		if err != nil {
			if err == io.EOF {
				log.Log().Infof("Peer closed connection: %s", peer.node)
			} else {
				log.Log().Errorf("Peer connection error: %s", peer.node, err)
			}
			n.removePeer(peer)
			return
		}
		// TODO: Handle message
		log.Log().Infof("Received message from peer: %s", msg.String())
	}
}

func (n *p2pNetwork) removePeer(peer *peer) {
	n.peers[peer.node.ID] = nil
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
