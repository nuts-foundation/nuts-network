package nodelist

import (
	"encoding/json"
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/pkg/documentlog"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/p2p"
	"time"
)

const nodeInfoDocumentType = "node-info"

type nodeList struct {
	documents             documentlog.DocumentLog
	p2pNetwork            p2p.P2PNetwork
	publicAddr            string
}

func (n *nodeList) Start(nodeID model.NodeID, publicAddr string) {
	// TODO: Test if already registered by examining local copy to avoid creating waste (duplicate documents) on the network.
	//  If not in the local copy, we should not autoregister our node but this should be done via a command.
	//  This process can be made easier by registering our node info when `registry register-vendor` is executed.
	if publicAddr != "" {
		n.publicAddr = publicAddr
		log.Log().Infof("Registering local node on nodelist (id=%s,addr=%s)", nodeID, publicAddr)
		data, _ := json.Marshal(model.NodeInfo{ID: nodeID, Address: publicAddr})
		n.documents.AddDocument(&model.Document{Timestamp: time.Now(), Contents: data, Type: nodeInfoDocumentType})
	}
	documentQueue := n.documents.Subscribe(nodeInfoDocumentType)
	go n.consumeNodeInfo(documentQueue)
}

func (n nodeList) Stop() {
	// TODO: stop loops
}

func (n *nodeList) consumeNodeInfo(queue documentlog.DocumentQueue) {
	// TODO: When to cancel?
	for {
		document := queue.Get()
		nodeInfo := model.NodeInfo{}
		if err := json.Unmarshal(document.Contents, &nodeInfo); err != nil {
			log.Log().Errorf("Can't parse node info from document (hash=%s)", document.Hash())
			continue
		}
		log.Log().Tracef("Parsed node info (ID=%s,Addr=%s)", nodeInfo.ID, nodeInfo.Address)
		n.p2pNetwork.AddRemoteNode(nodeInfo)
	}
}
