/*
 * Copyright (C) 2020. Nuts community
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 *
 */

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

