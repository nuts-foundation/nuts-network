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
	"errors"
	"github.com/golang/mock/gomock"
	cryptoMock "github.com/nuts-foundation/nuts-crypto/test/mock"
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-network/pkg/documentlog"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/nodelist"
	"github.com/nuts-foundation/nuts-network/pkg/p2p"
	"github.com/nuts-foundation/nuts-network/pkg/proto"
	"github.com/nuts-foundation/nuts-network/pkg/stats"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

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
		cxt := createNetwork(ctrl)
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
		cxt := createNetwork(ctrl)
		cxt.documentLog.EXPECT().GetDocument(gomock.Any())
		cxt.network.GetDocument(model.EmptyHash())
	})
}

func TestNetwork_GetDocumentContents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	t.Run("ok", func(t *testing.T) {
		cxt := createNetwork(ctrl)
		cxt.documentLog.EXPECT().GetDocumentContents(gomock.Any())
		cxt.network.GetDocumentContents(model.EmptyHash())
	})
}

func TestNetwork_Diagnostics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	t.Run("ok", func(t *testing.T) {
		cxt := createNetwork(ctrl)
		cxt.documentLog.EXPECT().Statistics().Return([]stats.Statistic{stat{}, stat{}})
		cxt.protocol.EXPECT().Statistics().Return([]stats.Statistic{stat{}, stat{}})
		cxt.p2pNetwork.EXPECT().Statistics().Return([]stats.Statistic{stat{}, stat{}})
		diagnostics := cxt.network.Diagnostics()
		assert.Len(t, diagnostics, 6)
	})
}

func TestNetwork_Start(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	t.Run("ok - client mode", func(t *testing.T) {
		cxt := createNetwork(ctrl)
		err := cxt.network.Start()
		if !assert.NoError(t, err) {
			return
		}
	})
	t.Run("ok - server mode", func(t *testing.T) {
		cxt := createNetwork(ctrl)

		cxt.p2pNetwork.EXPECT().Start()
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
		cxt := createNetwork(ctrl)
		cxt.network.Config.Mode = core.ClientEngineMode
		err := cxt.network.Shutdown()
		assert.NoError(t, err)
	})
	t.Run("ok - server mode", func(t *testing.T) {
		cxt := createNetwork(ctrl)
		cxt.documentLog.EXPECT().Stop()
		cxt.nodeList.EXPECT().Stop()
		cxt.p2pNetwork.EXPECT().Stop()
		cxt.network.Config.Mode = core.ServerEngineMode
		err := cxt.network.Shutdown()
		assert.NoError(t, err)
	})
	t.Run("error - server mode, stop returns error", func(t *testing.T) {
		cxt := createNetwork(ctrl)
		cxt.documentLog.EXPECT().Stop()
		cxt.nodeList.EXPECT().Stop()
		cxt.p2pNetwork.EXPECT().Stop().Return(errors.New("failed"))
		cxt.network.Config.Mode = core.ServerEngineMode
		err := cxt.network.Shutdown()
		assert.EqualError(t, err, "failed")
	})
}

func TestNetwork_buildP2PNetworkConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	t.Run("ok", func(t *testing.T) {
		cxt := createNetwork(ctrl)
		cxt.network.Config.CertFile = "../test-files/certificate-and-key.pem"
		cxt.network.Config.CertKeyFile = "../test-files/certificate-and-key.pem"
		cxt.crypto.EXPECT().TrustStore()
		cfg, err := cxt.network.buildP2PConfig()
		assert.NotNil(t, cfg)
		assert.NoError(t, err)
		assert.NotNil(t, cfg.ClientCert)
		assert.Equal(t, cfg.ClientCert, cfg.ServerCert)
	})
	t.Run("error - unable to load key pair from file", func(t *testing.T) {
		cxt := createNetwork(ctrl)
		cxt.network.Config.CertFile = "../test-files/non-existent.pem"
		cxt.crypto.EXPECT().TrustStore()
		cfg, err := cxt.network.buildP2PConfig()
		assert.Nil(t, cfg)
		assert.EqualError(t, err, "unable to load node TLS certificate and/or key from file: open ../test-files/non-existent.pem: no such file or directory")
	})
	t.Run("error - unable to load key pair from crypto", func(t *testing.T) {
		cxt := createNetwork(ctrl)
		cxt.crypto.EXPECT().TrustStore()
		cxt.crypto.EXPECT().GetTLSCertificate(gomock.Any()).Return(nil, nil, errors.New("failed"))
		cfg, err := cxt.network.buildP2PConfig()
		assert.Nil(t, cfg)
		assert.EqualError(t, err, "unable to load node TLS certificate and/or key from crypto module: failed")
	})
}

func createNetwork(ctrl *gomock.Controller) *networkTestContext {
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
	}
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
