package proto

import (
	"github.com/magiconair/properties/assert"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"sync"
	"testing"
)

func TestPeerConsistencyHashDiagnostic(t *testing.T) {
	diagnostic := peerConsistencyHashDiagnostic{peerHashes: new(map[model.PeerID]model.Hash), mux: &sync.Mutex{}}
	diagnostic.copyFrom(map[model.PeerID]model.Hash{"abc": []byte{1, 2, 3}})
	assert.Equal(t, diagnostic.String(), "010203={abc}")
}
