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
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"google.golang.org/grpc/metadata"
	"net"
	"strings"
)

const nodeIDHeader = "nodeID"

func normalizeAddress(addr string) string {
	var normalizedAddr string
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		normalizedAddr = addr
	} else {
		if host == "localhost" {
			host = "127.0.0.1"
			normalizedAddr = net.JoinHostPort(host, port)
		} else {
			normalizedAddr = addr
		}
	}
	return normalizedAddr
}

func nodeIDFromMetadata(md metadata.MD) (model.NodeID, error) {
	values := md.Get(nodeIDHeader)
	if len(values) == 0 {
		return "", fmt.Errorf("peer didn't send %s header", nodeIDHeader)
	} else if len(values) > 1 {
		return "", fmt.Errorf("peer sent multiple values for %s header", nodeIDHeader)
	}
	nodeID := model.NodeID(strings.TrimSpace(values[0]))
	if nodeID.Empty() {
		return "", fmt.Errorf("peer sent empty %s header", nodeIDHeader)
	}
	return nodeID, nil
}

func constructMetadata(nodeID model.NodeID) metadata.MD {
	return metadata.New(map[string]string{nodeIDHeader: string(nodeID)})
}
