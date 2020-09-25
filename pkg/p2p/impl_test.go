package p2p

import (
	"crypto/tls"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	"github.com/nuts-foundation/nuts-go-test/io"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Test_p2pNetwork_Configure(t *testing.T) {
	t.Run("ok - configure registers bootstrap nodes", func(t *testing.T) {
		network := NewP2PNetwork()
		ts, _ := cert.NewTrustStore(filepath.Join(os.TempDir(), "truststore.pem.test"))
		err := network.Configure(P2PNetworkConfig{
			NodeID:         "foo",
			ListenAddress:  "0.0.0.0:555",
			BootstrapNodes: []string{"foo:555", "bar:5554"},
			TrustStore:     ts,
		})
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, network.(*p2pNetwork).remoteNodeAddChannel, 2)
	})
}

func Test_p2pNetwork_Start(t *testing.T) {
	waitForGRPCStart := func() {
		time.Sleep(100 * time.Millisecond) // Wait a moment for gRPC server setup goroutines to run
	}
	t.Run("ok - gRPC server not bound", func(t *testing.T) {
		network := NewP2PNetwork().(*p2pNetwork)
		ts, _ := cert.NewTrustStore(filepath.Join(os.TempDir(), "truststore.pem.test"))
		err := network.Configure(P2PNetworkConfig{
			NodeID:     "foo",
			TrustStore: ts,
		})
		if !assert.NoError(t, err) {
			return
		}
		err = network.Start()
		waitForGRPCStart()
		assert.Nil(t, network.listener)
		defer network.Stop()
		if !assert.NoError(t, err) {
			return
		}
	})
	t.Run("ok - gRPC server bound, TLS enabled", func(t *testing.T) {
		network := NewP2PNetwork().(*p2pNetwork)
		ts, _ := cert.NewTrustStore(filepath.Join(os.TempDir(), "truststore.pem.test"))
		serverCert, _ := tls.LoadX509KeyPair("../../test-files/certificate-and-key.pem", "../../test-files/certificate-and-key.pem")
		err := network.Configure(P2PNetworkConfig{
			NodeID:        "foo",
			ServerCert:    serverCert,
			ListenAddress: ":5555",
			TrustStore:    ts,
		})
		if !assert.NoError(t, err) {
			return
		}
		err = network.Start()
		waitForGRPCStart()
		assert.NotNil(t, network.listener)
		defer network.Stop()
		if !assert.NoError(t, err) {
			return
		}
	})
	t.Run("ok - gRPC server bound, TLS disabled", func(t *testing.T) {
		testDirectory := io.TestDirectory(t)
		network := NewP2PNetwork().(*p2pNetwork)
		ts, _ := cert.NewTrustStore(filepath.Join(testDirectory, "truststore.pem"))
		err := network.Configure(P2PNetworkConfig{
			NodeID:        "foo",
			ListenAddress: ":5555",
			TrustStore:    ts,
		})
		if !assert.NoError(t, err) {
			return
		}
		err = network.Start()
		waitForGRPCStart()
		assert.NotNil(t, network.listener)
		defer network.Stop()
		if !assert.NoError(t, err) {
			return
		}
	})
}

func Test_p2pNetwork_GetLocalAddress(t *testing.T) {
	network := NewP2PNetwork().(*p2pNetwork)
	ts, _ := cert.NewTrustStore(filepath.Join(os.TempDir(), "truststore.pem.test"))
	err := network.Configure(P2PNetworkConfig{
		NodeID:         "foo",
		ListenAddress:  "0.0.0.0:555",
		BootstrapNodes: []string{"foo:555", "bar:5554"},
		TrustStore:     ts,
	})
	if !assert.NoError(t, err) {
		return
	}
	t.Run("ok - public address not configured, listen address fully qualified", func(t *testing.T) {
		assert.Equal(t, "0.0.0.0:555", network.getLocalAddress())
	})
	t.Run("ok - public address not configured, listen address contains only port", func(t *testing.T) {
		network.config.ListenAddress = ":555"
		assert.Equal(t, "localhost:555", network.getLocalAddress())
	})
	t.Run("ok - public address configured", func(t *testing.T) {
		network.config.PublicAddress = "test:1234"
		assert.Equal(t, "test:1234", network.getLocalAddress())
	})
}
