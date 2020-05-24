package p2p

import (
	"encoding/hex"
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/network"
	grpcPeer "google.golang.org/grpc/peer"
	"io"
)

func (n p2pNetwork) Connect(stream network.Network_ConnectServer) error {
	peerCtx, _ := grpcPeer.FromContext(stream.Context())
	log.Log().Infof("New peer connected from %s", peerCtx.Addr)
	for {
		message, err := stream.Recv()
		if err == io.EOF {
			log.Log().Info("EOF detected, peer disconnected.")
			// TODO: More work to be done here?
			return nil
		} else if err != nil {
			log.Log().Warn("gRPC error occurred: ", err)
			return err
		}

		if message.Header == nil {
			// TODO: Close stream, or does it happen auto?
			return ErrMissingProtocolVersion
		} else if message.Header.Version != ProtocolVersion {
			// TODO: Close stream, or does it happen auto?
			return ErrUnsupportedProtocolVersion
		}

		if message.AdvertLastHash != nil {
			log.Log().Info("Received last hash from peer: " + hex.EncodeToString(message.AdvertLastHash.Hash))
		}
	}
}

