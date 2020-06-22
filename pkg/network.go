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

package pkg

import (
	"bytes"
	"crypto/tls"
	crypto "github.com/nuts-foundation/nuts-crypto/pkg"
	"github.com/nuts-foundation/nuts-crypto/pkg/types"
	core "github.com/nuts-foundation/nuts-go-core"
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/pkg/documentlog"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/nodelist"
	"github.com/nuts-foundation/nuts-network/pkg/p2p"
	"github.com/nuts-foundation/nuts-network/pkg/proto"
	"github.com/nuts-foundation/nuts-network/pkg/stats"
	errors2 "github.com/pkg/errors"
	"io"
	"strings"
	"sync"
	"time"
)

// ModuleName defines the name of this module
const ModuleName = "Network"

// NetworkConfig holds the config
type NetworkConfig struct {
	// Socket address for gRPC to listen on
	GrpcAddr string
	Mode     string
	Address  string
	// Public address of this nodes other nodes can use to connect to this node.
	PublicAddr     string
	BootstrapNodes string
	NodeID         string
	CertFile       string
	CertKeyFile    string
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
	crypto      crypto.Client
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

// Configure configures the network subsystem
func (n *Network) Configure() error {
	var err error
	n.configOnce.Do(func() {
		cfg := core.NutsConfig()
		n.Config.Mode = cfg.GetEngineMode(n.Config.Mode)
		if n.Config.Address == "" {
			n.Config.Address = cfg.ServerAddress()
		}
		if n.Config.Mode == core.ServerEngineMode {
			n.crypto = crypto.NewCryptoClient()
			for _, nodeInfo := range n.Config.ParseBootstrapNodes() {
				n.p2pNetwork.AddRemoteNode(model.ParseNodeInfo(nodeInfo))
			}
		}
	})
	return err
}

// Start initiates the network subsystem
func (n *Network) Start() error {
	if n.Config.Mode != core.ServerEngineMode {
		return nil
	}
	networkConfig, err := n.buildP2PConfig()
	if err != nil {
		return err
	}
	if err := n.p2pNetwork.Start(*networkConfig); err != nil {
		return err
	}
	n.protocol.Start(n.p2pNetwork, n.documentLog)
	n.documentLog.Start()
	n.nodeList.Start(networkConfig.NodeID, networkConfig.PublicAddress)
	return nil
}

func (n *Network) GetDocument(hash model.Hash) (*model.Document, error) {
	return n.documentLog.GetDocument(hash), nil
}

func (n *Network) GetDocumentContents(hash model.Hash) (io.ReadCloser, error) {
	return n.documentLog.GetDocumentContents(hash)
}

func (n *Network) ListDocuments() ([]model.Document, error) {
	return n.documentLog.Documents(), nil
}

func (n *Network) AddDocumentWithContents(timestamp time.Time, docType string, contents []byte) (model.Document, error) {
	log.Log().Infof("Adding document (timestamp=%d,type=%s,content length=%d)", timestamp.UnixNano(), docType, len(contents))
	// TODO: Validation
	return n.documentLog.AddDocumentWithContents(timestamp, docType, bytes.NewReader(contents))
}

// Shutdown cleans up any leftover go routines
func (n *Network) Shutdown() error {
	if n.Config.Mode == core.ServerEngineMode {
		n.nodeList.Stop()
		n.documentLog.Stop()
		return n.p2pNetwork.Stop()
	}
	return nil
}

func (n *Network) Diagnostics() []core.DiagnosticResult {
	var statistics = make([]stats.Statistic, 0)
	statistics = append(statistics, n.documentLog.Statistics()...)
	statistics = append(statistics, n.protocol.Statistics()...)
	statistics = append(statistics, n.p2pNetwork.Statistics()...)
	var result = make([]core.DiagnosticResult, len(statistics))
	// Hey that's convenient, stats.Statistic matches core.DiagnosticResult!
	for i, statistic := range statistics {
		result[i] = statistic
	}
	return result
}

func (n *Network) buildP2PConfig() (*p2p.P2PNetworkConfig, error) {
	cfg := p2p.P2PNetworkConfig{
		ListenAddress: n.Config.GrpcAddr,
		PublicAddress: n.Config.PublicAddr,
		NodeID:        model.NodeID(n.Config.NodeID),
		TrustStore:    n.crypto.TrustStore(),
	}
	identity := core.NutsConfig().Identity()
	if cfg.NodeID == "" {
		log.Log().Warn("NodeID not configured, will use node identity.")
		cfg.NodeID = model.NodeID(identity)
	}
	if n.Config.CertFile == "" && n.Config.CertKeyFile == "" {
		log.Log().Info("No certificate and/or key file specified, will load TLS certificate from crypto module.")
		tlsCertificate, privateKey, err := n.crypto.GetTLSCertificate(types.KeyForEntity(types.LegalEntity{URI: identity}))
		if err != nil {
			return nil, errors2.Wrap(err, "unable to load node TLS certificate and/or key from crypto module")
		}
		cfg.ServerCert = tls.Certificate{
			Certificate: [][]byte{tlsCertificate.Raw},
			PrivateKey:  privateKey,
			Leaf:        tlsCertificate,
		}
	} else {
		log.Log().Info("Will load TLS certificate from specified certificate/key file")
		var err error
		if cfg.ServerCert, err = tls.LoadX509KeyPair(n.Config.CertFile, n.Config.CertKeyFile); err != nil {
			return nil, errors2.Wrap(err, "unable to load node TLS certificate and/or key from file")
		}
	}
	cfg.ClientCert = cfg.ServerCert
	return &cfg, nil
}
