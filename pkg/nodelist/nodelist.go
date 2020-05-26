package nodelist

import (
	"encoding/json"
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/pkg/doclog"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"time"
)

// NodeList is one of the applications built on top of the Nuts network which is used for discovering new nodes.
type NodeList interface {
	Start(nodeId model.NodeID, address string)
	Stop()
}

func NewNodeList(log doclog.DocumentLog) NodeList {
	return &nodeList{documents: log}
}

type nodeList struct {
	documents doclog.DocumentLog
}

func (n nodeList) Start(nodeId model.NodeID, address string) {
	// TODO: Test if already registered
	log.Log().Infof("Registering local node on nodelist (id=%s,addr=%s)", nodeId, address)
	contents := map[string]string{
		"id":   nodeId.String(),
		"addr": address,
	}
	data, _ := json.Marshal(contents)
	n.documents.Add(&model.Document{Timestamp: time.Now(), Contents: data})
}

func (n nodeList) Stop() {
	// TODO
}
