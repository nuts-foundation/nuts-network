package p2p

import (
	"fmt"
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/network"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"google.golang.org/grpc"
	"io"
)

type messageGate interface {
	Send(message *network.NetworkMessage) error
	Recv() (*network.NetworkMessage, error)
}

// Peer represents a connected peer
type peer struct {
	id     model.PeerID
	nodeID model.NodeID
	addr   string

	// gate is used to send and receive messages
	gate messageGate
	// conn and client are only filled for peers where we're the connecting party
	conn   *grpc.ClientConn
	client network.NetworkClient
	// outMessages contains the messages we want to send to the peer.
	//   According to the docs it's unsafe to simultaneously call stream.Send() from multiple goroutines so we put them
	//   on a channel so that each peer can have its own goroutine sending messages (consuming messages from this channel)
	outMessages chan *network.NetworkMessage
}

func (p peer) String() string {
	return fmt.Sprintf("%s(%s)", p.nodeID, p.addr)
}

func (p *peer) close() {
	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			log.Log().Errorf("Unable to close client connection (peer=%s): %v", p, err)
		}
		p.conn = nil
	}
	close(p.outMessages)
}

func (p *peer) send(message *network.NetworkMessage) {
	p.outMessages <- message
}

// startSending (blocking) reads messages from its outMessages channel and sends them to the peer until the channel is closed.
func (p *peer) startSending() {
	p.outMessages = make(chan *network.NetworkMessage, 10) // TODO: Does this number make sense? Should also be configurable?
	for message := range p.outMessages {
		if p.gate.Send(message) != nil {
			log.Log().Errorf("Unable to broadcast message to peer (peer=%s)", p.id)
		}
	}
}

// startReceiving (blocking) reads messages from the peer until it disconnects or the network is stopped.
func (p *peer) startReceiving(peer *peer, queue messageQueue) {
	for {
		msg, recvErr := peer.gate.Recv()
		if recvErr != nil {
			if recvErr == io.EOF {
				log.Log().Infof("Peer closed connection: %s", peer)
			} else {
				log.Log().Errorf("Peer connection error (peer=%s): %v", peer, recvErr)
			}
			break
		}
		log.Log().Tracef("Received message from peer (%s): %s", peer, msg.String())
		queue.c <- PeerMessage{
			Peer:    peer.id,
			Message: msg,
		}
	}
}