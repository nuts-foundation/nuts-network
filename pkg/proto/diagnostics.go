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
