package nodelist

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_nodeList_Configure(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		err := (&nodeList{}).Configure("abc", "foo:8080")
		assert.NoError(t, err)
	})
	t.Run("error - nodeID empty", func(t *testing.T) {
		err := (&nodeList{}).Configure("", "foo:8080")
		assert.EqualError(t, err, "nodeID is empty")
	})
	t.Run("ok - address empty", func(t *testing.T) {
		err := (&nodeList{}).Configure("abc", "")
		assert.NoError(t, err)
	})
}