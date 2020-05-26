package p2p

import (
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/network"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"io"
)

// receiveFromPeer reads messages from the peer until it disconnects or the network is stopped.
func (n p2pNetwork) receiveFromPeer(peer *peer, gate messageGate) error {
	var err error
	for n.isRunning() {
		msg, recvErr := gate.Recv()
		if recvErr != nil {
			if err == io.EOF {
				log.Log().Infof("Peer closed connection: %s", peer)
			} else {
				log.Log().Errorf("Peer connection error (peer=%s): %v", peer, recvErr)
			}
			break
		}
		log.Log().Debugf("Received message from peer (%s): %s", peer, msg.String())
		if msg.Header == nil {
			err = ErrMissingProtocolVersion
			break
		} else if msg.Header.Version != ProtocolVersion {
			err = ErrUnsupportedProtocolVersion
			break
		}
		if err = n.handleMessageContents(peer, msg, gate); err != nil {
			// Should we do this? This disconnects the stream, while it is an internal error on our side
			break
		}
	}
	if err != nil {
		log.Log().Errorf("Protocol error occurred, closing connection (peer=%s): %v", peer, err)
	}
	// TODO: Synchronization
	n.removePeer(peer)
	return err
}

func (n p2pNetwork) handleMessageContents(peer *peer, msg *network.NetworkMessage, gate messageGate) error {
	if msg.AdvertHash != nil {
		log.Log().Debugf("Received adverted hash from peer: %s", peer)
		n.receivedConsistencyHashes.put(PeerHash{
			Peer: peer.id,
			Hash: msg.AdvertHash.Hash,
		})
	}
	if msg.HashListQuery != nil {
		log.Log().Infof("Received hash list query from peer, responding with consistency hash list (peer=%s)", peer)
		if n.hashSource == nil {
			log.Log().Error("Hash source hasn't been set, can't respond to query!")
			return nil
		}
		consistencyHashes := n.hashSource.ConsistencyHashes()
		msg := createMessage()
		msg.HashList = &network.HashList{
			Hash: make([][]byte, len(consistencyHashes)),
		}
		for i, consistencyHash := range consistencyHashes {
			msg.HashList.Hash[i] = consistencyHash
		}
		if err := gate.Send(&msg); err != nil {
			return err
		}
	}
	if msg.HashList != nil {
		log.Log().Infof("Received hash list from peer (peer=%s)", peer)
		for _, hash := range msg.HashList.Hash {
			log.Log().Infof(" %s", model.Hash(hash))
		}
	}
	return nil
}
