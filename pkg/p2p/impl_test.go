package p2p

import (
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
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
