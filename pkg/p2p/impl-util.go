package p2p

import (
	"fmt"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"google.golang.org/grpc/metadata"
	"net"
	"strings"
)

const NodeIDHeader = "nodeID"

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
	vals := md.Get(NodeIDHeader)
	if len(vals) == 0 {
		return "", fmt.Errorf("peer didn't send %s header", NodeIDHeader)
	} else if len(vals) > 1 {
		return "", fmt.Errorf("peer sent multiple values for %s header", NodeIDHeader)
	}
	return model.NodeID(strings.TrimSpace(vals[0])), nil
}

func constructMetadata(nodeID model.NodeID) metadata.MD {
	return metadata.New(map[string]string{NodeIDHeader: string(nodeID)})
}
