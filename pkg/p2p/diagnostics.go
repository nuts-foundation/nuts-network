package p2p

import (
	"strings"
)

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
	return strings.Join(addrs, " ")
}

