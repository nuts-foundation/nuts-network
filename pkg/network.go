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
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/p2p"
	"strings"
	"sync"
)

// ModuleName defines the name of this module
const ModuleName = "Network"

// NetworkClient is the interface to be implemented by any remote or local client
type NetworkClient interface {
}

// NetworkConfig holds the config
type NetworkConfig struct {
	GrpcAddr       string
	BootstrapNodes string
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
	Config     NetworkConfig
	configOnce sync.Once
	p2pNetwork p2p.P2PNetwork
}

var instance *Network
var oneRegistry sync.Once

// NetworkInstance returns the singleton Network
func NetworkInstance() *Network {
	oneRegistry.Do(func() {
		instance = &Network{
			p2pNetwork: p2p.NewP2PNetwork(),
		}
	})

	return instance
}

// Configure initializes the db, but only when in server mode
func (n *Network) Configure() error {
	var err error

	n.configOnce.Do(func() {
		for _, nodeInfo := range n.Config.ParseBootstrapNodes() {
			n.p2pNetwork.AddRemoteNode(model.ParseNodeInfo(nodeInfo))
		}
	})
	return err
}

// Start initiates the network subsystem
func (n *Network) Start() error {
	networkConfig := p2p.P2PNetworkConfig{ListenAddress: n.Config.GrpcAddr}
	return n.p2pNetwork.Start(networkConfig)
}

// Shutdown cleans up any leftover go routines
func (n *Network) Shutdown() error {
	return n.p2pNetwork.Stop()
}
