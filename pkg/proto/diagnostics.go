package proto

import (
	"fmt"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"strings"
)

type peerConsistencyHashDiagnostic struct {
	peerHashes map[model.PeerID]model.Hash
}

func (p peerConsistencyHashDiagnostic) Name() string {
	return "[Protocol] Peer hashes"
}

func (p peerConsistencyHashDiagnostic) String() string {
	var groupedByHash = make(map[string][]string)
	for peer, hash := range p.peerHashes {
		groupedByHash[hash.String()] = append(groupedByHash[hash.String()], string(peer))
	}
	var items []string
	for hash, peers := range groupedByHash {
		items = append(items, fmt.Sprintf("%s={%s}", hash, strings.Join(peers, ", ")))
	}
	return strings.Join(items, ", ")
}
