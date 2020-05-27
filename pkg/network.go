/*
 * Nuts registry
 * Copyright (C) 2019. Nuts community
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

package pkg

import (
	"errors"
	core "github.com/nuts-foundation/nuts-go-core"
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/pkg/documentlog"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/nodelist"
	"github.com/nuts-foundation/nuts-network/pkg/p2p"
	"github.com/nuts-foundation/nuts-network/pkg/proto"
	"github.com/spf13/cobra"
	"strings"
	"sync"
	"time"
)

// ModuleName defines the name of this module
const ModuleName = "Network"

// NetworkClient is the interface to be implemented by any remote or local client
type NetworkClient interface {
}

// NetworkConfig holds the config
type NetworkConfig struct {
	// Socket address for gRPC to listen on
	GrpcAddr string
	// Public address of this nodes other nodes can use to connect to this node.
	PublicAddr     string
	BootstrapNodes string
	NodeID         string
}

func (c NetworkConfig) ParseBootstrapNodes() []string {
	var result []string
	for _, addr := range strings.Split(c.BootstrapNodes, " ") {
		trimmedAddr := strings.TrimSpace(addr)
		if trimmedAddr != "" {
			result = append(result, trimmedAddr)
		}
	}
	return result
}

// Network holds the config and Db reference
type Network struct {
	Config      NetworkConfig
	configOnce  sync.Once
	p2pNetwork  p2p.P2PNetwork
	protocol    proto.Protocol
	documentLog documentlog.DocumentLog
	nodeList    nodelist.NodeList
}

var instance *Network
var oneRegistry sync.Once

// NetworkInstance returns the singleton Network
func NetworkInstance() *Network {
	oneRegistry.Do(func() {
		instance = &Network{
			p2pNetwork: p2p.NewP2PNetwork(),
			protocol:   proto.NewProtocol(),
		}
		instance.documentLog = documentlog.NewDocumentLog(instance.protocol)
		instance.nodeList = nodelist.NewNodeList(instance.documentLog, instance.p2pNetwork)
	})

	return instance
}

// Configure initializes the db, but only when in server mode
func (n *Network) Configure() error {
	var err error

	n.configOnce.Do(func() {
		// TODO: Why isn't this in core?
		core.NutsConfig().Load(&cobra.Command{})
		if core.NutsConfig().Identity() == "" {
			err = errors.New("nuts identity not configured")
			return
		}
		for _, nodeInfo := range n.Config.ParseBootstrapNodes() {
			n.p2pNetwork.AddRemoteNode(model.ParseNodeInfo(nodeInfo))
		}
	})
	return err
}

// Start initiates the network subsystem
func (n *Network) Start() error {
	networkConfig := p2p.P2PNetworkConfig{
		ListenAddress: n.Config.GrpcAddr,
		PublicAddress: n.Config.PublicAddr,
		NodeID:        model.NodeID(n.Config.NodeID),
	}
	if networkConfig.NodeID == "" {
		log.Log().Warn("NodeID not configured, will use node identity.")
		networkConfig.NodeID = model.NodeID(core.NutsConfig().Identity())
	}
	if err := n.p2pNetwork.Start(networkConfig); err != nil {
		return err
	}
	n.protocol.Start(n.p2pNetwork, n.documentLog)
	n.documentLog.Start()
	n.nodeList.Start(networkConfig.NodeID, networkConfig.PublicAddress)
	return nil
}

func (n *Network) AddDocument(contents []byte) error {
	n.documentLog.AddDocument(&model.Document{
		Contents:  contents,
		Timestamp: time.Now(),
	})
	return nil
}

// Shutdown cleans up any leftover go routines
func (n *Network) Shutdown() error {
	n.nodeList.Stop()
	n.documentLog.Stop()
	return n.p2pNetwork.Stop()
}

func (n *Network) Diagnostics() []core.DiagnosticResult {
	var result = make([]core.DiagnosticResult, 0)
	result = append(result, n.documentLog.Diagnostics()...)
	result = append(result, n.protocol.Diagnostics()...)
	result = append(result, n.p2pNetwork.Diagnostics()...)
	return result
}
