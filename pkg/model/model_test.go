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

func TestMakeConsistencyHash(t *testing.T) {
	t.Run("ok - h1 empty", func(t *testing.T) {
		expected, _ := ParseHash("383c9da631bd120169e82b0679e4c2e8d5050894")
		actual := MakeConsistencyHash(EmptyHash(), expected)
		assert.Equal(t, expected.String(), actual.String())
	})
	t.Run("ok - both empty", func(t *testing.T) {
		expected := EmptyHash()
		actual := MakeConsistencyHash(EmptyHash(), expected)
		assert.Equal(t, expected.String(), actual.String())
	})
}