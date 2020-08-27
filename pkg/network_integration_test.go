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
	"context"
	"fmt"
	crypto "github.com/nuts-foundation/nuts-crypto/pkg"
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-network/pkg/documentlog"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/nodelist"
	"github.com/nuts-foundation/nuts-network/pkg/p2p"
	"github.com/nuts-foundation/nuts-network/pkg/proto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"sync"
	"testing"
	"time"
)

const bufSize = 1024 * 1024
const defaultTimeout = 2 * time.Second
const documentType = "test"

var bufconListeners = make(map[string]*bufconn.Listener, 0)
var mutex = sync.Mutex{}

func TestNetwork(t *testing.T) {
	documentlog.AdvertHashInterval = 500 * time.Millisecond
	expectedDocLogSize := 0

	// Start 3 nodes: bootstrap, node1 and node2. Node 1 and 2 connect to the bootstrap node and should discover
	// each other that way.
	bootstrap, err := startNode("bootstrap")
	expectedDocLogSize++ // node registration
	if !assert.NoError(t, err) {
		return
	}
	node1, err := startNode("node1")
	expectedDocLogSize++ // node registration
	if !assert.NoError(t, err) {
		return
	}
	node1.p2pNetwork.AddRemoteNode(model.NodeInfo{
		ID:      "bootstrap",
		Address: "bootstrap",
	})
	node2, err := startNode("node2")
	expectedDocLogSize++ // node registration
	if !assert.NoError(t, err) {
		return
	}
	node2.p2pNetwork.AddRemoteNode(model.NodeInfo{
		ID:      "bootstrap",
		Address: "bootstrap",
	})
	stop := func() {
		node2.Shutdown()
		node1.Shutdown()
		bootstrap.Shutdown()
	}

	//time.Sleep(1000 * time.Second)
	// Wait until nodes are connected
	if !waitFor(t, func() (bool, error) {
		return len(node1.p2pNetwork.Peers()) == 2 && len(node2.p2pNetwork.Peers()) == 2, nil
	}, defaultTimeout) {
		stop()
		return
	}

	// Add a document on node1 and we expect in to come out on node2
	if addDocumentAndWaitForItToArrive(t, node1, node2) {
		stop()
		return
	}
	expectedDocLogSize++
	// Add a document on node2 and we expect in to come out on node1
	if addDocumentAndWaitForItToArrive(t, node2, node1) {
		stop()
		return
	}
	expectedDocLogSize++

	// Now assert that all nodes have received all documents
	waitForDocuments := func(docLog documentlog.DocumentLog) bool {
		return waitFor(t, func() (bool, error) {
			if docs, err := docLog.Documents(); err != nil {
				return false, err
			} else {
				return len(docs) == expectedDocLogSize, nil
			}
		}, defaultTimeout)
	}
	waitForDocuments(bootstrap.documentLog)
	waitForDocuments(node1.documentLog)
	waitForDocuments(node2.documentLog)

	// Can we request the diagnostics?
	fmt.Printf("%v\n", bootstrap.Diagnostics())
	fmt.Printf("%v\n", node1.Diagnostics())
	fmt.Printf("%v\n", node2.Diagnostics())
}

func addDocumentAndWaitForItToArrive(t *testing.T, sender *Network, receiver *Network) bool {
	receiverSub := receiver.documentLog.Subscribe(documentType)
	addedDocument, err := sender.AddDocumentWithContents(time.Now(), documentType, []byte("foobar"))
	if !assert.NoError(t, err) {
		return true
	}
	receivedDocument := receiverSub.Get()
	addedDocument.Timestamp = addedDocument.Timestamp.UTC()
	receivedDocument.Timestamp = receivedDocument.Timestamp.UTC()
	assert.Equal(t, addedDocument, receivedDocument)
	return false
}

func startNode(name string) (*Network, error) {
	mutex.Lock()
	bufconListeners[name] = bufconn.Listen(bufSize)
	mutex.Unlock()
	cryptoInstance := crypto.CryptoInstance()
	cryptoInstance.Config = crypto.CryptoConfig{
		Keysize: crypto.MinKeySize,
		Fspath:  "../test-files",
	}
	logrus.SetLevel(logrus.DebugLevel)
	if err := cryptoInstance.Configure(); err != nil {
		return nil, err
	}
	instance = &Network{
		p2pNetwork: p2p.NewP2PNetworkWithOptions(bufconListeners[name], func(ctx context.Context, target string, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error) {
			dialer := grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
				return bufconListeners[target].Dial()
			})
			mutex.Lock()
			defer mutex.Unlock()
			return grpc.DialContext(ctx, target, dialer, grpc.WithBlock(), grpc.WithInsecure())
		}),
		protocol: proto.NewProtocol(),
		crypto:   cryptoInstance,
		Config: NetworkConfig{
			GrpcAddr:                name,
			Mode:                    core.ServerEngineMode,
			StorageConnectionString: "file:" + name + "?mode=memory&cache=shared&busy_timeout=500",
			PublicAddr:              name,
			BootstrapNodes:          "",
			NodeID:                  name,
			CertFile:                "../test-files/certificate-and-key.pem",
			CertKeyFile:             "../test-files/certificate-and-key.pem",
		},
	}
	instance.documentLog = documentlog.NewDocumentLog(instance.protocol)
	instance.nodeList = nodelist.NewNodeList(instance.documentLog, instance.p2pNetwork)
	if err := instance.Configure(); err != nil {
		return nil, err
	}
	if err := instance.Start(); err != nil {
		return nil, err
	}
	return instance, nil
}

type predicate func() (bool, error)

func waitFor(t *testing.T, p predicate, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for {
		b, err := p()
		if !assert.NoError(t, err) {
			return false
		}
		if b {
			return true
		}
		if time.Now().After(deadline) {
			assert.Fail(t, "wait timeout")
			return false
		}
		time.Sleep(100 * time.Millisecond)
	}
}
