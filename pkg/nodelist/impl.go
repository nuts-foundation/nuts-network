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
	"bytes"
	"encoding/json"
	"errors"
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/pkg/documentlog"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/p2p"
	"time"
)

const nodeInfoDocumentType = "nuts.node-info"

type nodeList struct {
	documentLog documentlog.DocumentLog
	p2pNetwork  p2p.P2PNetwork
	localNode   model.NodeInfo
}

func (n *nodeList) Stop() {

}

func (n *nodeList) Configure(nodeID model.NodeID, address string) error {
	if nodeID.Empty() {
		return errors.New("nodeID is empty")
	}
	if address == "" {
		return errors.New("address is empty")
	}
	n.localNode = model.NodeInfo{
		ID:      nodeID,
		Address: address,
	}
	return nil
}

func (n *nodeList) Start() {
	// TODO: Test if already registered by examining local copy to avoid creating waste (duplicate documents) on the network.
	//  If not in the local copy, we should not autoregister our node but this should be done via a command.
	//  This process can be made easier by registering our node info when `registry register-vendor` is executed.
	if n.localNode.Address != "" {
		log.Log().Infof("Registering local node on nodelist: %s", n.localNode)
		data, _ := json.Marshal(n.localNode)
		if _, err := n.documentLog.AddDocumentWithContents(time.Now(), nodeInfoDocumentType, bytes.NewReader(data)); err != nil {
			// TODO: Shouldn't this be blocking?
			log.Log().Errorf("Error while adding document with contents: %v", err)
		}
	}
	go n.consumeNodeInfoFromQueue(n.documentLog.Subscribe(nodeInfoDocumentType))
}

func (n *nodeList) consumeNodeInfoFromQueue(queue documentlog.DocumentQueue) {
	for {
		document := queue.Get()
		if document == nil {
			return
		}
		if err := n.consumeNodeInfo(*document); err != nil {
			log.Log().Errorf("Error while processing nodelist document (hash=%s): %v", document.Hash, err)
		}
	}
}

func (n *nodeList) consumeNodeInfo(document model.Document) error {
	nodeInfo := model.NodeInfo{}
	reader, err := n.documentLog.GetDocumentContents(document.Hash)
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
