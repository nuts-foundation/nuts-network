package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDocument_Clone(t *testing.T) {
	t.Run("ok - empty hash", func(t *testing.T) {
		hash, _ := ParseHash("37076f2cbe014a109d79b61ae9c7191f4cc57afc")
		expected := Document{
			Hash:      hash,
			Type:      "foobar",
			Timestamp: time.Now(),
		}
		actual := expected.Clone()
		assert.Equal(t, expected, actual)
		assert.NotSame(t, &expected, &actual)
		assert.NotSame(t, expected.Hash, actual.Hash)
	})
	t.Run("ok - assert hash is copied", func(t *testing.T) {
		hash, _ := ParseHash("37076f2cbe014a109d79b61ae9c7191f4cc57afc")
		expected := Document{
			Hash:      hash,
			Type:      "foobar",
			Timestamp: time.Now(),
		}
		actual := expected.Clone()
		copy(expected.Hash, EmptyHash())
		assert.NotEqual(t, expected, actual)
	})
	t.Run("ok - nil hash", func(t *testing.T) {
		expected := Document{
			Hash:      nil,
			Type:      "foobar",
			Timestamp: time.Now(),
		}
		actual := expected.Clone()
		assert.Equal(t, expected, actual)
		assert.NotSame(t, &expected, &actual)
		assert.NotSame(t, expected.Hash, actual.Hash)
	})
}