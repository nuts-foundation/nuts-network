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
	"bytes"
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
		p.peerConsistencyHashDiagnostic,
	}
}

func NewProtocol() Protocol {
	return &protocol{
		receivedConsistencyHashes: &AdvertedHashQueue{
			c: make(chan PeerHash, 100), // TODO: Does this number make sense?
		},
		receivedDocumentHashes: &AdvertedHashQueue{
			c: make(chan PeerHash, 1000), // TODO: Does this number make sense?
		},

		peerHashes:         make(map[model.PeerID]model.Hash),
		newPeerHashChannel: make(chan PeerHash, 100),

		peerConsistencyHashDiagnostic: peerConsistencyHashDiagnostic{peerHashes: map[model.PeerID]model.Hash{}},
	}
}

func (p *protocol) Start(p2pNetwork p2p.P2PNetwork, hashSource HashSource) {
	p.p2pNetwork = p2pNetwork
	p.hashSource = hashSource
	go p.consumeMessages()
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
				p.peerConsistencyHashDiagnostic.peerHashes = copyPeerHashMap(p.peerHashes)
			}
		case peerHash := <-p.newPeerHashChannel:
			p.peerHashes[peerHash.Peer] = peerHash.Hash
			p.peerConsistencyHashDiagnostic.peerHashes = copyPeerHashMap(p.peerHashes)
		}
	}
}

func copyPeerHashMap(input map[model.PeerID]model.Hash) map[model.PeerID]model.Hash {
	var output = make(map[model.PeerID]model.Hash, len(input))
	for k, v := range input {
		output[k] = v
	}
	return output
}

func (p protocol) consumeMessages() {
	queue := p.p2pNetwork.ReceivedMessages()
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
		log.Log().Debugf("Received adverted hash from peer: %s", peer)
		peerHash := PeerHash{
			Peer: peer,
			Hash: msg.AdvertHash.Hash,
		}
		p.newPeerHashChannel <- peerHash
		p.receivedConsistencyHashes.c <- peerHash
	}
	if msg.HashListQuery != nil {
		log.Log().Debugf("Received hash list query from peer, responding with consistency hash list (peer=%s)", peer)
		msg := createMessage()
		documents := p.hashSource.Documents()
		msg.HashList = &network.HashList{
			Hashes: make([]*network.Document, len(documents)),
		}
		for i, hash := range documents {
			msg.HashList.Hashes[i] = &network.Document{
				Time: hash.Timestamp.UnixNano(),
				Hash: hash.Hash,
				Type: hash.Type,
			}
		}
		if err := p.p2pNetwork.Send(peer, &msg); err != nil {
			return err
		}
	}
	if msg.HashList != nil {
		log.Log().Debugf("Received hash list from peer (peer=%s)", peer)
		for _, current := range msg.HashList.Hashes {
			hash := model.Hash(current.Hash)
			if hash.Empty() {
				log.Log().Warn("Received document doesn't contain a hash, skipping.")
			}
			if current.Time == 0 {
				log.Log().Warnf("Received document doesn't contain a timestamp, skipping (hash=%s).", hash)
			}
			if !p.hashSource.HasDocument(hash) {
				document := model.Document{
					Type:      current.Type,
					Timestamp: time.Unix(0, current.Time),
					Hash:      current.Hash,
				}
				p.hashSource.AddDocument(document)
			}
			if !p.hashSource.HasContentsForDocument(hash) {
				// TODO: Currently we send the query to the peer that send us the hash, but this peer might not have the
				//   document contents. We need a smarter way to get it from a peer who does.
				log.Log().Infof("Received document hash from peer that we don't have yet, will query it (peer=%s,hash=%s,type=%s,timestamp=%d)", peer, hash, current.Type, current.Time)
				responseMsg := createMessage()
				responseMsg.DocumentContentsQuery = &network.DocumentContentsQuery{Hash: hash}
				if err := p.p2pNetwork.Send(peer, &responseMsg); err != nil {
					return err
				}
			}
		}
	}
	if msg.DocumentContentsQuery != nil && msg.DocumentContentsQuery.Hash != nil {
		hash := model.Hash(msg.DocumentContentsQuery.Hash)
		log.Log().Debugf("Received document contents query from peer (peer=%s, hash=%s)", peer, hash)
		// TODO: Maybe this should be asynchronous since loading document contents might be I/O heavy?
		if p.hashSource.HasContentsForDocument(hash) {
			reader, err := p.hashSource.GetDocumentContents(hash)
			responseMsg := createMessage()
			buffer := new(bytes.Buffer)
			_, err = buffer.ReadFrom(reader)
			if err != nil {
				log.Log().Warnf("Unable to read document contents (hash=%s): %v", hash, err)
			} else {
				responseMsg.DocumentContents = &network.DocumentContents{
					Hash:     hash,
					Contents: buffer.Bytes(),
				}
				if err := p.p2pNetwork.Send(peer, &responseMsg); err != nil {
					return err
				}
			}
		} else {
			log.Log().Warnf("Peer queried us for document contents, but we don't appear to have it (peer=%s,document=%s)", peer, hash)
		}
	}
	if msg.DocumentContents != nil && msg.DocumentContents.Hash != nil && msg.DocumentContents.Contents != nil {
		hash := model.Hash(msg.DocumentContents.Hash)
		log.Log().Infof("Received document contents from peer (peer=%s,hash=%s,len=%d)", peer, hash, len(msg.DocumentContents.Contents))
		// TODO: Maybe this should be asynchronous since writing the document contents might be I/O heavy?
		if !p.hashSource.HasDocument(hash) {
			log.Log().Warnf("We don't know the document we received contents for, ignoring (hash=%s)", hash)
		} else if p.hashSource.HasContentsForDocument(hash) {
			log.Log().Warnf("We already have the contents for the document, ignoring (hash=%s)", hash)
		} else {
			if err := p.hashSource.AddDocumentContents(hash, bytes.NewReader(msg.DocumentContents.Contents)); err != nil {
				log.Log().Errorf("Error while writing content for document (hash=%s): %v", hash, err)
			}
		}
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
