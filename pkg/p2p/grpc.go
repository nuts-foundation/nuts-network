package p2p

import (
	"encoding/hex"
	"errors"
	"fmt"
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/network"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	errors2 "github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
	grpcPeer "google.golang.org/grpc/peer"
	"io"
	"strings"
)

type messageGate interface {
	Send(message *network.NetworkMessage) error
	Recv() (*network.NetworkMessage, error)
}

func (n p2pNetwork) Connect(stream network.Network_ConnectServer) error {
	peerCtx, _ := grpcPeer.FromContext(stream.Context())
	log.Log().Infof("New peer connected from %s", peerCtx.Addr)
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return errors.New("unable to get metadata")
	}
	nodeId, err := nodeIdFromMetadata(md)
	if err != nil {
		return err
	}
	// We received our peer's NodeID, now send our own.
	if err := stream.SendHeader(constructMetadata(n.node.ID)); err != nil {
		return errors2.Wrap(err, "unable to send headers")
	}
	peer := &peer{
		// TODO
		id:     model.GetPeerID(peerCtx.Addr.String()),
		nodeId: nodeId,
		gate:   stream,
		addr:   peerCtx.Addr.String(),
	}
	n.addPeer(peer)
	return n.receiveFromPeer(peer, stream)
}

func nodeIdFromMetadata(md metadata.MD) (model.NodeID, error) {
	vals := md.Get(NodeIDHeader)
	if len(vals) == 0 {
		return "", fmt.Errorf("peer didn't send %s header", NodeIDHeader)
	} else if len(vals) > 1 {
		return "", fmt.Errorf("peer sent multiple values for %s header", NodeIDHeader)
	}
	return model.NodeID(strings.TrimSpace(vals[0])), nil
}

func constructMetadata(nodeId model.NodeID) metadata.MD {
	return metadata.New(map[string]string{NodeIDHeader: string(nodeId)})
}

// receiveFromPeer reads messages from the peer until it disconnects or the network is stopped.
func (n p2pNetwork) receiveFromPeer(peer *peer, gate messageGate) error {
	var err error
	for n.isRunning() {
		msg, recvErr := gate.Recv()
		if recvErr != nil {
			if err == io.EOF {
				log.Log().Infof("Peer closed connection: %s", peer)
			} else {
				log.Log().Errorf("Peer connection error: %s", peer, recvErr)
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
		if msg.AdvertLastHash != nil {
			log.Log().Infof("Received last hash from peer (%s): %s", peer, hex.EncodeToString(msg.AdvertLastHash.Hash))
		}
	}
	if err != nil {
		log.Log().Errorf("Protocol error occurred, closing connection: %s", peer, err)
	}
	// TODO: Synchronization
	n.removePeer(peer)
	return err
}
