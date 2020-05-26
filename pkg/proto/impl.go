package proto

import (
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/network"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/p2p"
	"time"
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
	// TODO: Everything works synchronous here, we might want to introduce some asynchronicity to avoid waiting on each other over the network
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
		log.Log().Debugf("Received hash list query from peer, responding with consistency hash list (peer=%s)", peer)
		msg := createMessage()
		msg.HashList = &network.HashList{
			Hashes: make([]*network.HashInTime, len(p.hashSource.DocumentHashes())),
		}
		for i, hash := range p.hashSource.DocumentHashes() {
			msg.HashList.Hashes[i] = &network.HashInTime{
				Time: hash.Timestamp.UnixNano(),
				Hash: hash.Hash,
			}
		}
		if err := p.p2pNetwork.Send(peer, &msg); err != nil {
			return err
		}
	}
	if msg.HashList != nil {
		log.Log().Debugf("Received hash list from peer (peer=%s)", peer)
		for _, hashInTime := range msg.HashList.Hashes {
			hash := model.Hash(hashInTime.Hash)
			p.hashSource.AddDocumentHash(hash, time.Unix(0, hashInTime.Time))
			if !p.hashSource.HasDocument(hash) {
				log.Log().Infof("Received document hash from peer that we don't have yet, will query it (peer=%s,hash=%s)", peer, model.Hash(hash))
				responseMsg := createMessage()
				responseMsg.DocumentQuery = &network.DocumentQuery{Hash: hash}
				if err := p.p2pNetwork.Send(peer, &responseMsg); err != nil {
					return err
				}
			}
		}
	}
	if msg.DocumentQuery != nil && msg.DocumentQuery.Hash != nil {
		log.Log().Debugf("Received document query from peer (peer=%s, hash=%s)", peer)
		hash := model.Hash(msg.DocumentQuery.Hash)
		document := p.hashSource.GetDocument(hash)
		if document == nil {
			log.Log().Warnf("Peer queried us for a document, but we don't appear to have it (peer=%s,document=%s)", peer, hash)
			return nil
		}
		responseMsg := createMessage()
		responseMsg.Document = &network.Document{
			Time:     document.Timestamp.UnixNano(),
			Contents: document.Contents,
		}
		if err := p.p2pNetwork.Send(peer, &responseMsg); err != nil {
			return err
		}
	}
	if msg.Document != nil && msg.Document.Contents != nil {
		log.Log().Infof("Received document from peer (peer=%s,time=%d)", peer, msg.Document.Time)
		p.hashSource.AddDocument(&model.Document{
			Contents:  msg.Document.Contents,
			Timestamp: time.Unix(0, msg.Document.Time),
		})
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
