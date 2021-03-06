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
	"fmt"
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/network"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"google.golang.org/grpc"
	"io"
	"sync"
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
	// closeMutex the close() function since race conditions can trigger panics
	closeMutex *sync.Mutex
}

func (p peer) String() string {
	return fmt.Sprintf("%s(%s)", p.nodeID, p.addr)
}

func (p *peer) close() {
	p.closeMutex.Lock()
	defer p.closeMutex.Unlock()
	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			log.Log().Errorf("Unable to close client connection (peer=%s): %v", p, err)
		}
		p.conn = nil
	}
	if p.outMessages != nil {
		close(p.outMessages)
		p.outMessages = nil
	}
}

func (p *peer) send(message *network.NetworkMessage) {
	p.closeMutex.Lock()
	defer p.closeMutex.Unlock()
	p.outMessages <- message
}

// sendMessages (blocking) reads messages from its outMessages channel and sends them to the peer until the channel is closed.
func (p peer) sendMessages() {
	for message := range p.outMessages {
		if p.gate.Send(message) != nil {
			log.Log().Errorf("Unable to broadcast message to peer (peer=%s)", p.id)
		}
	}
}

// receiveMessages (blocking) reads messages from the peer until it disconnects or the network is stopped.
func receiveMessages(gate messageGate, peerId model.PeerID, receivedMsgQueue messageQueue) {
	for {
		msg, recvErr := gate.Recv()
		if recvErr != nil {
			if recvErr == io.EOF {
				log.Log().Infof("Peer closed connection: %s", peerId)
			} else {
				log.Log().Errorf("Peer connection error (peer=%s): %v", peerId, recvErr)
			}
			break
		}
		log.Log().Tracef("Received message from peer (%s): %s", peerId, msg.String())
		receivedMsgQueue.c <- PeerMessage{
			Peer:    peerId,
			Message: msg,
		}
	}
}