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
