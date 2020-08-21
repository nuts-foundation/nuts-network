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

package nodelist

import (
	"github.com/nuts-foundation/nuts-network/pkg/documentlog"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/p2p"
)

func NewNodeList(log documentlog.DocumentLog, p2pNetwork p2p.P2PNetwork) NodeList {
	return &nodeList{documentLog: log, p2pNetwork: p2pNetwork}
}

// NodeList is one of the applications built on top of the Nuts network which is used for discovering new nodes.
type NodeList interface {
	// Configure configures the NodeList. Must be called before Start().
	Configure(nodeID model.NodeID, address string) error
	Start()
	Stop()
}
