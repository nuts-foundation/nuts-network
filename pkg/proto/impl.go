package proto

import (
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/network"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/p2p"
)

type protocol struct {
	p2pNetwork p2p.P2PNetwork
	hashSource HashSource
	// TODO: What if no-one is actually listening to this queue? Maybe we should create it when someone asks for it (lazy initialization)?
	receivedConsistencyHashes *AdvertedHashQueue
	receivedDocumentHashes    *AdvertedHashQueue
}

func NewProtocol() Protocol {
	return &protocol{
		receivedConsistencyHashes: &AdvertedHashQueue{
			c: make(chan PeerHash, 100), // TODO: Does this number make sense?
		},
		receivedDocumentHashes: &AdvertedHashQueue{
			c: make(chan PeerHash, 1000), // TODO: Does this number make sense?
		},
	}
}

func (p *protocol) Start(p2pNetwork p2p.P2PNetwork, hashSource HashSource) {
	p.p2pNetwork = p2pNetwork
	p.hashSource = hashSource
	go p.consumeMessages()
}

func (p protocol) Stop() {
	panic("implement me")
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

func (p protocol) handleMessage(peerMsg p2p.PeerMessage) error {
	peer := peerMsg.Peer
	msg := peerMsg.Message
	if msg.AdvertHash != nil {
		log.Log().Debugf("Received adverted hash from peer: %s", peer)
		p.receivedConsistencyHashes.put(PeerHash{
			Peer: peer,
			Hash: msg.AdvertHash.Hash,
		})
	}
	if msg.HashListQuery != nil {
		log.Log().Infof("Received hash list query from peer, responding with consistency hash list (peer=%s)", peer)
		consistencyHashes := p.hashSource.ConsistencyHashes()
		msg := createMessage()
		msg.HashList = &network.HashList{
			Hash: make([][]byte, len(consistencyHashes)),
		}
		for i, consistencyHash := range consistencyHashes {
			msg.HashList.Hash[i] = consistencyHash
		}
		if err := p.p2pNetwork.Send(peer, &msg); err != nil {
			return err
		}
	}
	if msg.HashList != nil {
		log.Log().Infof("Received hash list from peer (peer=%s)", peer)
		for _, hash := range msg.HashList.Hash {
			if !p.hashSource.ContainsDocument(hash) {
				log.Log().Infof("Received document hash from peer that we don't have yet, will query it (peer=%s,hash=%s)", peer, hash)
				queryMsg := createMessage()
				queryMsg.
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
