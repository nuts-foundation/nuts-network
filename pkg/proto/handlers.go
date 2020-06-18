package proto

import (
	"bytes"
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/network"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"time"
)

func (p *protocol) handleAdvertHash(peer model.PeerID, advertHash *network.AdvertHash) {
	log.Log().Debugf("Received adverted hash from peer: %s", peer)
	peerHash := PeerHash{
		Peer: peer,
		Hash: advertHash.Hash,
	}
	p.newPeerHashChannel <- peerHash
	p.receivedConsistencyHashes.internal.Add(peerHash)
}

func (p *protocol) handleDocumentContents(peer model.PeerID, contents *network.DocumentContents) {
	hash := model.Hash(contents.Hash)
	log.Log().Infof("Received document contents from peer (peer=%s,hash=%s,len=%d)", peer, hash, len(contents.Contents))
	// TODO: Maybe this should be asynchronous since writing the document contents might be I/O heavy?
	if _, err := p.hashSource.AddDocumentContents(hash, bytes.NewReader(contents.Contents)); err != nil {
		log.Log().Errorf("Error while writing content for document (hash=%s): %v", hash, err)
	}
}

func (p *protocol) handleDocumentContentsQuery(peer model.PeerID, query *network.DocumentContentsQuery) error {
	hash := model.Hash(query.Hash)
	log.Log().Debugf("Received document contents query from peer (peer=%s, hash=%s)", peer, hash)
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
				Hash:     hash,
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
	log.Log().Debugf("Received hash list from peer (peer=%s)", peer)
	var documents = make([]model.Document, len(hashList.Hashes))
	for i, current := range hashList.Hashes {
		hash := model.Hash(current.Hash)
		if hash.Empty() {
			log.Log().Warn("Received document doesn't contain a hash, skipping.")
		}
		if current.Time == 0 {
			log.Log().Warnf("Received document doesn't contain a timestamp, skipping (hash=%s).", hash)
		}
		documents[i] = model.Document{
			Type:      current.Type,
			Timestamp: time.Unix(0, current.Time),
			Hash:      current.Hash,
		}
	}
	missingContentHashes, err := p.hashSource.AddMissingDocuments(documents)
	if err != nil {
		return err
	}
	if len(missingContentHashes) > 0 {
		// TODO: Currently we send the query to the peer that send us the hash, but this peer might not have the
		//   document contents. We need a smarter way to get it from a peer who does.
		log.Log().Infof("Document hashes from peer that we don't have yet, will query it (peer=%s,hashes=%v)", peer, missingContentHashes)
		for _, hash := range missingContentHashes {
			responseMsg := createMessage()
			responseMsg.DocumentContentsQuery = &network.DocumentContentsQuery{Hash: hash}
			if err := p.p2pNetwork.Send(peer, &responseMsg); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *protocol) handleHashListQuery(peer model.PeerID) error {
	log.Log().Debugf("Received hash list query from peer, responding with consistency hash list (peer=%s)", peer)
	msg := createMessage()
	documents, err := p.hashSource.Documents()
	if err != nil {
		return err
	}
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
	return nil
}