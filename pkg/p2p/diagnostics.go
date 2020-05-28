package p2p

import (
	"fmt"
	"sort"
	"strings"
)

type NumberOfDiagnosticsResult struct {
	NumberOfPeers int
}

func (n NumberOfDiagnosticsResult) Name() string {
	return "[P2P Network] Connected peers #"
}

func (n NumberOfDiagnosticsResult) String() string {
	return fmt.Sprintf("%d", n.NumberOfPeers)
}

type PeersDiagnosticsResult struct {
	Peers []Peer
}

func (p PeersDiagnosticsResult) Name() string {
	return "[P2P Network] Connected peers"
}

func (p PeersDiagnosticsResult) String() string {
	addrs := make([]string, len(p.Peers))
	for i, peer := range p.Peers {
		addrs[i] = peer.String()
	}
	// Sort for stable order (easier for humans to understand)
	sort.Slice(addrs, func(i, j int) bool {
		return addrs[i] > addrs[j]
	})
	return strings.Join(addrs, " ")
}

