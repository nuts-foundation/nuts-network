package p2p

import (
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
	"testing"
)

func Test_normalizeAddress(t *testing.T) {
	t.Run("invalid address", func(t *testing.T) {
		assert.Equal(t, "not a valid address", normalizeAddress("not a valid address"))
	})
	t.Run("address is already normalized (IP)", func(t *testing.T) {
		assert.Equal(t, "1.2.3.4:1234", normalizeAddress("1.2.3.4:1234"))
	})
	t.Run("address is already normalized (hostname)", func(t *testing.T) {
		assert.Equal(t, "foobar:1234", normalizeAddress("foobar:1234"))
	})
	t.Run("address is localhost (hostname)", func(t *testing.T) {
		assert.Equal(t, "127.0.0.1:1234", normalizeAddress("localhost:1234"))
	})
	t.Run("address is localhost (IP)", func(t *testing.T) {
		assert.Equal(t, "127.0.0.1:1234", normalizeAddress("127.0.0.1:1234"))
	})
}

func Test_nodeIDFromMetadata(t *testing.T) {
	t.Run("ok - roundtrip", func(t *testing.T) {
		md := constructMetadata("1234")
		nodeId, err := nodeIDFromMetadata(md)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "1234", nodeId.String())
	})
	t.Run("error - multiple values", func(t *testing.T) {
		md := metadata.MD{}
		md.Append(nodeIDHeader, "1")
		md.Append(nodeIDHeader, "2")
		nodeId, err := nodeIDFromMetadata(md)
		assert.EqualError(t, err, "peer sent multiple values for nodeID header")
		assert.Empty(t, nodeId.String())
	})
	t.Run("error - no values", func(t *testing.T) {
		md := metadata.MD{}
		nodeId, err := nodeIDFromMetadata(md)
		assert.EqualError(t, err, "peer didn't send nodeID header")
		assert.Empty(t, nodeId.String())
	})
	t.Run("error - empty value", func(t *testing.T) {
		md := metadata.MD{}
		md.Set(nodeIDHeader, "  ")
		nodeId, err := nodeIDFromMetadata(md)
		assert.EqualError(t, err, "peer sent empty nodeID header")
		assert.Empty(t, nodeId.String())
	})
}