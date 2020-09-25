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
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"github.com/golang/mock/gomock"
	cryptoMock "github.com/nuts-foundation/nuts-crypto/test/mock"
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-go-test/io"
	"github.com/nuts-foundation/nuts-network/pkg/documentlog"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/nodelist"
	"github.com/nuts-foundation/nuts-network/pkg/p2p"
	"github.com/nuts-foundation/nuts-network/pkg/proto"
	"github.com/nuts-foundation/nuts-network/pkg/stats"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

const vendorID = "urn:oid:1.3.6.1.4.1.54851.4:vendorId"

type networkTestContext struct {
	network     Network
	documentLog *documentlog.MockDocumentLog
	nodeList    *nodelist.MockNodeList
	p2pNetwork  *p2p.MockP2PNetwork
	protocol    *proto.MockProtocol
	crypto      *cryptoMock.MockClient
}

func TestNetwork_ListDocuments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	t.Run("ok", func(t *testing.T) {
		cxt := createNetwork(t, ctrl)
		cxt.documentLog.EXPECT().Documents().Return([]model.DocumentDescriptor{{
			Document: model.Document{
				Hash:      model.EmptyHash(),
				Type:      "foo",
				Timestamp: time.Now(),
			},
		}}, nil)
		docs, err := cxt.network.ListDocuments()
		assert.Len(t, docs, 1)
		assert.NoError(t, err)
	})
}

func TestNetwork_GetDocument(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	t.Run("ok", func(t *testing.T) {
		cxt := createNetwork(t, ctrl)
		cxt.documentLog.EXPECT().GetDocument(gomock.Any())
		cxt.network.GetDocument(model.EmptyHash())
	})
}

func TestNetwork_GetDocumentContents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	t.Run("ok", func(t *testing.T) {
		cxt := createNetwork(t, ctrl)
		cxt.documentLog.EXPECT().GetDocumentContents(gomock.Any())
		cxt.network.GetDocumentContents(model.EmptyHash())
	})
}

func TestNetwork_Subscribe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	t.Run("ok", func(t *testing.T) {
		cxt := createNetwork(t, ctrl)
		cxt.documentLog.EXPECT().Subscribe("some-type")
		cxt.network.Subscribe("some-type")
	})
}

func TestNetwork_Diagnostics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	t.Run("ok", func(t *testing.T) {
		cxt := createNetwork(t, ctrl)
		cxt.documentLog.EXPECT().Statistics().Return([]stats.Statistic{stat{}, stat{}})
		cxt.protocol.EXPECT().Statistics().Return([]stats.Statistic{stat{}, stat{}})
		cxt.p2pNetwork.EXPECT().Statistics().Return([]stats.Statistic{stat{}, stat{}})
		diagnostics := cxt.network.Diagnostics()
		assert.Len(t, diagnostics, 6)
	})
}

func TestNetwork_Configure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	t.Run("ok", func(t *testing.T) {
		cxt := createNetwork(t, ctrl)
		cxt.protocol.EXPECT().Configure(gomock.Any(), gomock.Any())
		cxt.documentLog.EXPECT().Configure(gomock.Any())
		cxt.nodeList.EXPECT().Configure(model.NodeID(vendorID), "foo:8080")
		cxt.p2pNetwork.EXPECT().Configure(gomock.Any())
		cxt.crypto.EXPECT().TrustStore()
		cxt.crypto.EXPECT().GetTLSCertificate(gomock.Any()).Return(&x509.Certificate{}, &rsa.PrivateKey{}, nil)
		err := cxt.network.Configure()
		if !assert.NoError(t, err) {
			return
		}
	})
	t.Run("ok - renew tls cert", func(t *testing.T) {
		cxt := createNetwork(t, ctrl)
		cxt.protocol.EXPECT().Configure(gomock.Any(), gomock.Any())
		cxt.documentLog.EXPECT().Configure(gomock.Any())
		cxt.nodeList.EXPECT().Configure(gomock.Any(), gomock.Any())
		cxt.p2pNetwork.EXPECT().Configure(gomock.Any())
		cxt.crypto.EXPECT().TrustStore()
		cxt.crypto.EXPECT().GetTLSCertificate(gomock.Any()).Return(nil, nil, nil)
		cxt.crypto.EXPECT().RenewTLSCertificate(gomock.Any()).Return(&x509.Certificate{}, &rsa.PrivateKey{}, nil)
		err := cxt.network.Configure()
		if !assert.NoError(t, err) {
			return
		}
	})
	t.Run("error - unable to renew tls cert", func(t *testing.T) {
		cxt := createNetwork(t, ctrl)
		cxt.protocol.EXPECT().Configure(gomock.Any(), gomock.Any())
		cxt.documentLog.EXPECT().Configure(gomock.Any())
		cxt.nodeList.EXPECT().Configure(gomock.Any(), gomock.Any())
		cxt.crypto.EXPECT().TrustStore()
		cxt.crypto.EXPECT().GetTLSCertificate(gomock.Any()).Return(nil, nil, nil)
		cxt.crypto.EXPECT().RenewTLSCertificate(gomock.Any()).Return(nil, nil, errors.New("failed"))
		err := cxt.network.Configure()
		assert.NoError(t, err)
	})
	t.Run("ok - no TLS certificate, network offline", func(t *testing.T) {
		cxt := createNetwork(t, ctrl)
		cxt.protocol.EXPECT().Configure(gomock.Any(), gomock.Any())
		cxt.documentLog.EXPECT().Configure(gomock.Any())
		cxt.nodeList.EXPECT().Configure(gomock.Any(), gomock.Any())
		cxt.crypto.EXPECT().TrustStore()
		cxt.crypto.EXPECT().GetTLSCertificate(gomock.Any()).Return(nil, nil, errors.New("failed"))
		err := cxt.network.Configure()
		assert.NoError(t, err)
	})
	t.Run("error - can't configure nodelist", func(t *testing.T) {
		cxt := createNetwork(t, ctrl)
		cxt.protocol.EXPECT().Configure(gomock.Any(), gomock.Any())
		cxt.documentLog.EXPECT().Configure(gomock.Any())
		cxt.nodeList.EXPECT().Configure(gomock.Any(), gomock.Any()).Return(errors.New("failed"))
		err := cxt.network.Configure()
		assert.EqualError(t, err, "unable to configure nodelist: failed")
	})
}

func TestNetwork_Start(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	t.Run("ok - client mode", func(t *testing.T) {
		cxt := createNetwork(t, ctrl)
		err := cxt.network.Start()
		if !assert.NoError(t, err) {
			return
		}
	})
	t.Run("ok - server mode", func(t *testing.T) {
		cxt := createNetwork(t, ctrl)
		cxt.p2pNetwork.EXPECT().Start()
		cxt.p2pNetwork.EXPECT().Configured().Return(true)
		cxt.protocol.EXPECT().Start()
		cxt.documentLog.EXPECT().Start()
		cxt.nodeList.EXPECT().Start()
		cxt.network.Config.Mode = core.ServerEngineMode
		err := cxt.network.Start()
		if !assert.NoError(t, err) {
			return
		}
	})
	t.Run("ok - server mode - network offline", func(t *testing.T) {
		cxt := createNetwork(t, ctrl)
		cxt.p2pNetwork.EXPECT().Configured().Return(false)
		cxt.protocol.EXPECT().Start()
		cxt.documentLog.EXPECT().Start()
		cxt.nodeList.EXPECT().Start()
		cxt.network.Config.Mode = core.ServerEngineMode
		err := cxt.network.Start()
		if !assert.NoError(t, err) {
			return
		}
	})
}

func TestNetwork_Shutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	t.Run("ok - CLI mode", func(t *testing.T) {
		cxt := createNetwork(t, ctrl)
		cxt.network.Config.Mode = core.ClientEngineMode
		err := cxt.network.Shutdown()
		assert.NoError(t, err)
	})
	t.Run("ok - server mode", func(t *testing.T) {
		cxt := createNetwork(t, ctrl)
		cxt.documentLog.EXPECT().Stop()
		cxt.nodeList.EXPECT().Stop()
		cxt.p2pNetwork.EXPECT().Stop()
		cxt.network.Config.Mode = core.ServerEngineMode
		err := cxt.network.Shutdown()
		assert.NoError(t, err)
	})
	t.Run("error - server mode, stop returns error", func(t *testing.T) {
		cxt := createNetwork(t, ctrl)
		cxt.documentLog.EXPECT().Stop()
		cxt.nodeList.EXPECT().Stop()
		cxt.p2pNetwork.EXPECT().Stop().Return(errors.New("failed"))
		cxt.network.Config.Mode = core.ServerEngineMode
		err := cxt.network.Shutdown()
		assert.EqualError(t, err, "failed")
	})
}

func TestDefaultNetworkConfig(t *testing.T) {
	defs := DefaultNetworkConfig()
	assert.True(t, defs.EnableTLS)
	assert.Equal(t, "file:network.db", defs.StorageConnectionString)
}

func TestNetwork_buildP2PNetworkConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	t.Run("ok - TLS enabled", func(t *testing.T) {
		cxt := createNetwork(t, ctrl)
		cxt.network.Config.GrpcAddr = ":5555"
		cxt.network.Config.EnableTLS = true
		cxt.network.Config.CertFile = "../test-files/certificate-and-key.pem"
		cxt.network.Config.CertKeyFile = "../test-files/certificate-and-key.pem"
		cxt.crypto.EXPECT().TrustStore()
		cxt.crypto.EXPECT().GetTLSCertificate(gomock.Any()).Return(&x509.Certificate{}, &rsa.PrivateKey{}, nil)
		cfg, err := cxt.network.buildP2PConfig()
		assert.NotNil(t, cfg)
		assert.NoError(t, err)
		assert.NotNil(t, cfg.ClientCert.PrivateKey)
		assert.NotNil(t, cfg.ServerCert.PrivateKey)
	})
	t.Run("ok - TLS disabled", func(t *testing.T) {
		cxt := createNetwork(t, ctrl)
		cxt.network.Config.GrpcAddr = ":5555"
		cxt.network.Config.EnableTLS = false
		cxt.crypto.EXPECT().TrustStore()
		cxt.crypto.EXPECT().GetTLSCertificate(gomock.Any()).Return(&x509.Certificate{}, &rsa.PrivateKey{}, nil)
		cfg, err := cxt.network.buildP2PConfig()
		assert.NotNil(t, cfg)
		assert.NoError(t, err)
		assert.NotNil(t, cfg.ClientCert.PrivateKey)
		assert.Nil(t, cfg.ServerCert.PrivateKey)
	})
	t.Run("ok - gRPC server not bound", func(t *testing.T) {
		cxt := createNetwork(t, ctrl)
		cxt.network.Config.GrpcAddr = ""
		cxt.network.Config.EnableTLS = true
		cxt.crypto.EXPECT().TrustStore()
		cxt.crypto.EXPECT().GetTLSCertificate(gomock.Any()).Return(&x509.Certificate{}, &rsa.PrivateKey{}, nil)
		cfg, err := cxt.network.buildP2PConfig()
		assert.NotNil(t, cfg)
		assert.NoError(t, err)
		assert.NotNil(t, cfg.ClientCert.PrivateKey)
		assert.Nil(t, cfg.ServerCert.PrivateKey)
	})
	t.Run("error - unable to load key pair from file", func(t *testing.T) {
		cxt := createNetwork(t, ctrl)
		cxt.network.Config.CertFile = "../test-files/non-existent.pem"
		cxt.network.Config.CertKeyFile = "../test-files/non-existent.pem"
		cxt.network.Config.EnableTLS = true
		cxt.crypto.EXPECT().TrustStore()
		cxt.crypto.EXPECT().GetTLSCertificate(gomock.Any()).Return(&x509.Certificate{}, &rsa.PrivateKey{}, nil)
		cfg, err := cxt.network.buildP2PConfig()
		assert.Nil(t, cfg)
		assert.EqualError(t, err, "unable to load node TLS server certificate and/or key from file: open ../test-files/non-existent.pem: no such file or directory")
	})
	t.Run("error - TLS enabled, certificate key file not configured", func(t *testing.T) {
		cxt := createNetwork(t, ctrl)
		cxt.network.Config.GrpcAddr = ":5555"
		cxt.network.Config.EnableTLS = true
		cxt.network.Config.CertFile = "../test-files/certificate-and-key.pem"
		cxt.network.Config.CertKeyFile = ""
		cxt.crypto.EXPECT().TrustStore()
		cxt.crypto.EXPECT().GetTLSCertificate(gomock.Any()).Return(&x509.Certificate{}, &rsa.PrivateKey{}, nil)
		cfg, err := cxt.network.buildP2PConfig()
		assert.Nil(t, cfg)
		assert.EqualError(t, err, "certFile and certKeyFile must be configured when enableTLS=true")
	})
	t.Run("error - unable to load key pair from crypto", func(t *testing.T) {
		cxt := createNetwork(t, ctrl)
		cxt.crypto.EXPECT().TrustStore()
		cxt.crypto.EXPECT().GetTLSCertificate(gomock.Any()).Return(nil, nil, errors.New("failed"))
		cfg, err := cxt.network.buildP2PConfig()
		assert.Nil(t, cfg)
		assert.EqualError(t, err, "unable to load node TLS client certificate and/or key from crypto module: failed")
	})
}

func createNetwork(t *testing.T, ctrl *gomock.Controller) *networkTestContext {
	os.Setenv("NUTS_IDENTITY", vendorID)
	core.NutsConfig().Load(&cobra.Command{})
	documentLog := documentlog.NewMockDocumentLog(ctrl)
	nodeList := nodelist.NewMockNodeList(ctrl)
	p2pNetwork := p2p.NewMockP2PNetwork(ctrl)
	protocol := proto.NewMockProtocol(ctrl)
	cryptoClient := cryptoMock.NewMockClient(ctrl)
	network := Network{
		documentLog: documentLog,
		nodeList:    nodeList,
		p2pNetwork:  p2pNetwork,
		protocol:    protocol,
		crypto:      cryptoClient,
		Config:      TestNetworkConfig(io.TestDirectory(t)),
	}
	network.Config.PublicAddr = "foo:8080"
	return &networkTestContext{
		network:     network,
		documentLog: documentLog,
		nodeList:    nodeList,
		p2pNetwork:  p2pNetwork,
		protocol:    protocol,
		crypto:      cryptoClient,
	}
}

type stat struct {
}

func (s stat) Name() string {
	panic("implement me")
}

func (s stat) String() string {
	panic("implement me")
}
