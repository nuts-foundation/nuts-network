package nodelist

import (
	"bytes"
	"encoding/json"
	"errors"
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/pkg/documentlog"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/p2p"
	"time"
)

const nodeInfoDocumentType = "node-info"

type nodeList struct {
	documents  documentlog.DocumentLog
	p2pNetwork p2p.P2PNetwork
	publicAddr string
}

func (n *nodeList) Start(nodeID model.NodeID, publicAddr string) {
	// TODO: Test if already registered by examining local copy to avoid creating waste (duplicate documents) on the network.
	//  If not in the local copy, we should not autoregister our node but this should be done via a command.
	//  This process can be made easier by registering our node info when `registry register-vendor` is executed.
	if publicAddr != "" {
		n.publicAddr = publicAddr
		log.Log().Infof("Registering local node on nodelist (id=%s,addr=%s)", nodeID, publicAddr)
		data, _ := json.Marshal(model.NodeInfo{ID: nodeID, Address: publicAddr})
		document := model.Document{
			Timestamp: time.Now(),
			Type:      nodeInfoDocumentType,
		}
		document.Hash = model.CalculateDocumentHash(document.Type, document.Timestamp, data)
		if err := n.documents.AddDocumentWithContents(document, bytes.NewReader(data)); err != nil {
			// TODO: Shouldn't this be blocking?
			log.Log().Errorf("Error while adding document with contents: %v", err)
		}
	}
	documentQueue := n.documents.Subscribe(nodeInfoDocumentType)
	go n.consumeNodeInfoFromQueue(documentQueue)
}

func (n nodeList) Stop() {
	// TODO: stop loops
}

func (n *nodeList) consumeNodeInfoFromQueue(queue documentlog.DocumentQueue) {
	// TODO: When to cancel?
	for {
		document := queue.Get()
		if err := n.consumeNodeInfo(document); err != nil {
			log.Log().Errorf("Error while processing nodelist document (hash=%s): %v", document.Hash, err)
		}
	}
}

func (n *nodeList) consumeNodeInfo(document model.Document) error {
	nodeInfo := model.NodeInfo{}
	reader, err := n.documents.GetDocumentContents(document.Hash)
	if err != nil {
		return err
	}
	if reader == nil {
		return errors.New("document contents not found")
	}
	defer reader.Close()
	buffer := new(bytes.Buffer)
	if _, err = buffer.ReadFrom(reader); err != nil {
		return err
	}
	if err := json.Unmarshal(buffer.Bytes(), &nodeInfo); err != nil {
		return err
	}
	log.Log().Tracef("Parsed node info (ID=%s,Addr=%s)", nodeInfo.ID, nodeInfo.Address)
	n.p2pNetwork.AddRemoteNode(nodeInfo)
	return nil
}