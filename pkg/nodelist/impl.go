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
	"github.com/nuts-foundation/nuts-network/pkg/documentlog/store"
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
	n.localNode = model.NodeInfo{
		ID:      nodeID,
		Address: address,
	}
	return nil
}

func (n *nodeList) Start() {
	if n.localNode.Address != "" {
		if err := n.register(); err != nil {
			log.Log().Errorf("Error while registering node: %v", err)
		}
	}
	go n.consumeNodeInfoFromQueue(n.documentLog.Subscribe(nodeInfoDocumentType))
}

// register registers this node's public address (node info) on the global node list, so other nodes can connect to this node.
// It only does so if the node info is not already present on the network to avoid producing duplicate data.
func (n *nodeList) register() error {
	data, _ := json.Marshal(n.localNode)
	if existingRegistrations, err := n.documentLog.FindByContentHash(store.HashContents(data)); err != nil {
		return err
	} else if len(existingRegistrations) == 0 {
		log.Log().Infof("Registering local node on nodelist: %s", n.localNode)
		if _, err := n.documentLog.AddDocumentWithContents(time.Now(), nodeInfoDocumentType, bytes.NewReader(data)); err != nil {
			return err
		}
	} else {
		log.Log().Info("Not registering local node on nodelist since node info is already present.")
	}
	return nil
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
