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

package proto

import (
	core "github.com/nuts-foundation/nuts-go-core"
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/network"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/p2p"
	"time"
)

// protocol is thread-safe when callers use the Protocol interface
type protocol struct {
	p2pNetwork p2p.P2PNetwork
	hashSource HashSource
	// TODO: What if no-one is actually listening to this queue? Maybe we should create it when someone asks for it (lazy initialization)?
	receivedConsistencyHashes *AdvertedHashQueue
	receivedDocumentHashes    *AdvertedHashQueue
	peerHashes                map[model.PeerID]model.Hash

	// Cache diagnostics to avoid having to lock precious resources
	peerConsistencyHashDiagnostic peerConsistencyHashDiagnostic
	newPeerHashChannel            chan PeerHash
}

func (p *protocol) Diagnostics() []core.DiagnosticResult {
	return []core.DiagnosticResult{
		&p.peerConsistencyHashDiagnostic,
	}
}

func NewProtocol() Protocol {
	p := &protocol{
		receivedConsistencyHashes: &AdvertedHashQueue{},
		receivedDocumentHashes:    &AdvertedHashQueue{},

		peerHashes:         make(map[model.PeerID]model.Hash),
		newPeerHashChannel: make(chan PeerHash, 100),

		peerConsistencyHashDiagnostic: newPeerConsistencyHashDiagnostic(),
	}
	// TODO: Does these numbers make sense?
	p.receivedConsistencyHashes.internal.Init(100)
	p.receivedDocumentHashes.internal.Init(1000)
	return p
}

func (p *protocol) Start(p2pNetwork p2p.P2PNetwork, hashSource HashSource) {
	p.p2pNetwork = p2pNetwork
	p.hashSource = hashSource
	go p.consumeMessages(p.p2pNetwork.ReceivedMessages())
	go p.updateDiagnostics()
}

func (p protocol) Stop() {

}

func (p protocol) ReceivedConsistencyHashes() PeerHashQueue {
	return p.receivedConsistencyHashes
}

func (p protocol) ReceivedDocumentHashes() PeerHashQueue {
	return p.receivedDocumentHashes
}

func (p protocol) AdvertConsistencyHash(hash model.Hash) {
	msg := createMessage()
	msg.AdvertHash = &network.AdvertHash{Hash: hash}
	p.p2pNetwork.Broadcast(&msg)
}

func (p protocol) QueryHashList(peer model.PeerID) error {
	msg := createMessage()
	msg.HashListQuery = &network.HashListQuery{}
	return p.p2pNetwork.Send(peer, &msg)
}

func (p *protocol) updateDiagnostics() {
	// TODO: When to exit the loop?
	ticker := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-ticker.C:
			// TODO: Make this garbage collection less dumb. Maybe we should be notified of disconnects rather than looping each time
			connectedPeers := p.p2pNetwork.Peers()
			var changed = false
			for peerId, _ := range p.peerHashes {
				var present = false
				for _, connectedPeer := range connectedPeers {
					if peerId == connectedPeer.PeerID {
						present = true
					}
				}
				if !present {
					delete(p.peerHashes, peerId)
					changed = true
				}
			}
			if changed {
				p.peerConsistencyHashDiagnostic.copyFrom(p.peerHashes)
			}
		case peerHash := <-p.newPeerHashChannel:
			p.peerHashes[peerHash.Peer] = peerHash.Hash
			p.peerConsistencyHashDiagnostic.copyFrom(p.peerHashes)
		}
	}
}

func (p protocol) consumeMessages(queue p2p.MessageQueue) {
	for {
		peerMsg := queue.Get()
		msg := peerMsg.Message
		var err error
		if msg.Header == nil {
			err = ErrMissingProtocolVersion
		} else if msg.Header.Version != Version {
			err = ErrUnsupportedProtocolVersion
		} else {
			err = p.handleMessage(peerMsg)
		}
		if err != nil {
			log.Log().Errorf("Error handling message (peer=%s): %v", peerMsg.Peer, err)
		}
	}
}

func (p *protocol) handleMessage(peerMsg p2p.PeerMessage) error {
	peer := peerMsg.Peer
	msg := peerMsg.Message
	if msg.AdvertHash != nil {
		p.handleAdvertHash(peer, msg.AdvertHash)
	}
	if msg.HashListQuery != nil {
		if err := p.handleHashListQuery(peer); err != nil {
			return err
		}
	}
	if msg.HashList != nil {
		if err := p.handleHashList(peer, msg.HashList); err != nil {
			return err
		}
	}
	if msg.DocumentContentsQuery != nil && msg.DocumentContentsQuery.Hash != nil {
		if err := p.handleDocumentContentsQuery(peer, msg.DocumentContentsQuery); err != nil {
			return err
		}
	}
	if msg.DocumentContents != nil && msg.DocumentContents.Hash != nil && msg.DocumentContents.Contents != nil {
		p.handleDocumentContents(peer, msg.DocumentContents)
	}
	return nil
}

func createMessage() network.NetworkMessage {
	return network.NetworkMessage{
		Header: &network.Header{
			Version: Version,
		},
	}
}
