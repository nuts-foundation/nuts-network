package nodelist

import (
	"github.com/nuts-foundation/nuts-network/pkg/documentlog"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/p2p"
)

func NewNodeList(log documentlog.DocumentLog, p2pNetwork p2p.P2PNetwork) NodeList {
	return &nodeList{documents: log, p2pNetwork: p2pNetwork}
}

// NodeList is one of the applications built on top of the Nuts network which is used for discovering new nodes.
type NodeList interface {
	Start(nodeID model.NodeID, address string)
	Stop()
}
