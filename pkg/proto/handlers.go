package proto

import (
	"bytes"
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/network"
	"github.com/nuts-foundation/nuts-network/pkg/model"
)

func (p *protocol) handleAdvertHash(peer model.PeerID, advertHash *network.AdvertHash) {
	log.Log().Tracef("Received adverted hash from peer: %s", peer)
	peerHash := PeerHash{
		Peer: peer,
		Hash: model.SliceToHash(advertHash.Hash),
	}
	p.newPeerHashChannel <- peerHash
	p.receivedConsistencyHashes.c <- &peerHash
}

func (p *protocol) handleDocumentContents(peer model.PeerID, contents *network.DocumentContents) {
	hash := model.SliceToHash(contents.Hash)
	log.Log().Infof("Received document contents from peer (peer=%s,hash=%s,len=%d)", peer, hash, len(contents.Contents))
	// TODO: Maybe this should be asynchronous since writing the document contents might be I/O heavy?
	if _, err := p.hashSource.AddDocumentContents(hash, bytes.NewReader(contents.Contents)); err != nil {
		log.Log().Errorf("Error while writing content for document (hash=%s): %v", hash, err)
	}
}

func (p *protocol) handleDocumentContentsQuery(peer model.PeerID, query *network.DocumentContentsQuery) error {
	hash := model.SliceToHash(query.Hash)
	log.Log().Tracef("Received document contents query from peer (peer=%s, hash=%s)", peer, hash)
	// TODO: Maybe this should be asynchronous since loading document contents might be I/O heavy?
	if contentsExist, err := p.hashSource.HasContentsForDocument(hash); err != nil {
		return err
	} else if contentsExist {
		reader, err := p.hashSource.GetDocumentContents(hash)
		responseMsg := createMessage()
		buffer := new(bytes.Buffer)
		_, err = buffer.ReadFrom(reader)
		if err != nil {
			log.Log().Warnf("Unable to read document contents (hash=%s): %v", hash, err)
		} else {
			responseMsg.DocumentContents = &network.DocumentContents{
				Hash:     query.Hash,
				Contents: buffer.Bytes(),
			}
			if err := p.p2pNetwork.Send(peer, &responseMsg); err != nil {
				return err
			}
		}
	} else {
		log.Log().Warnf("Peer queried us for document contents, test.appear to have it (peer=%s,document=%s)", peer, hash)
	}
	return nil
}

func (p *protocol) handleHashList(peer model.PeerID, hashList *network.HashList) error {
	log.Log().Tracef("Received hash list from peer (peer=%s)", peer)
	var documents = make([]model.Document, len(hashList.Hashes))
	for i, current := range hashList.Hashes {
		hash := model.SliceToHash(current.Hash)
		if hash.Empty() {
			log.Log().Warn("Received document doesn't contain a hash, skipping.")
			continue
		}
		if current.Time == 0 {
			log.Log().Warnf("Received document doesn't contain a timestamp, skipping (hash=%s).", hash)
			continue
		}
		documents[i] = model.Document{
			Type:      current.Type,
			Timestamp: model.UnmarshalDocumentTime(current.Time),
			Hash:      hash,
		}
	}
	for _, document := range documents {
		if p.documentIgnoreList[document.Hash] {
			continue
		}
		if err := p.checkDocumentOnLocalNode(peer, document); err != nil {
			log.Log().Errorf("Error while checking peer document on local node, ignoring it until next restart (peer=%s, document=%s): %v", peer, document.Hash, err)
			p.documentIgnoreList[document.Hash] = true
		}
	}
	return nil
}

func (p *protocol) checkDocumentOnLocalNode(peer model.PeerID, peerDocument model.Document) error {
	localDocument, err := p.hashSource.GetDocument(peerDocument.Hash)
	if err != nil {
		return err
	}
	if localDocument != nil && localDocument.HasContents {
		return nil
	} else if localDocument == nil {
		if err := p.hashSource.AddDocument(peerDocument); err != nil {
			return err
		}
	}
	// TODO: Currently we send the query to the peer that sent us the hash, but this peer might not have the
	//   document contents. We need a smarter way to get it from a peer who does.
	log.Log().Infof("Received document hash from peer that we don't have yet or we're missing its contents, will query it (peer=%s,hash=%s)", peer, peerDocument.Hash)
	responseMsg := createMessage()
	responseMsg.DocumentContentsQuery = &network.DocumentContentsQuery{Hash: peerDocument.Hash.Slice()}
	return p.p2pNetwork.Send(peer, &responseMsg)
}

func (p *protocol) handleHashListQuery(peer model.PeerID) error {
	log.Log().Tracef("Received hash list query from peer, responding with consistency hash list (peer=%s)", peer)
	msg := createMessage()
	documents, err := p.hashSource.Documents()
	if err != nil {
		return err
	}
	msg.HashList = &network.HashList{
		Hashes: make([]*network.Document, len(documents)),
	}
	for i, document := range documents {
		msg.HashList.Hashes[i] = &network.Document{
			Time: model.MarshalDocumentTime(document.Timestamp),
			Hash: document.Hash.Slice(),
			Type: document.Type,
		}
	}
	if err := p.p2pNetwork.Send(peer, &msg); err != nil {
		return err
	}
	return nil
}
