package proto

import (
	"fmt"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"sort"
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
		// Sort for stable order (easier for humans to understand)
		sort.Slice(peers, func(i, j int) bool {
			return peers[i] > peers[j]
		})
		items = append(items, fmt.Sprintf("%s={%s}", hash, strings.Join(peers, ", ")))
	}
	// Sort for stable order (easier for humans to understand)
	sort.Slice(items, func(i, j int) bool {
		return items[i] > items[j]
	})
	return strings.Join(items, ", ")
}
