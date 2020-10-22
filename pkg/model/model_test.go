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
		expected.Hash[0] = 0
		assert.NotEqual(t, expected, actual)
	})
	t.Run("ok - change type", func(t *testing.T) {
		unexpected := Document{
			Hash:      EmptyHash(),
			Type:      "foobar",
			Timestamp: time.Now(),
		}
		expected := unexpected.Clone()
		expected.Type = "test"
		assert.NotEqual(t, unexpected.Type, expected.Type)
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

func TestHash_Empty(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		assert.True(t, EmptyHash().Empty())
	})
	t.Run("non empty", func(t *testing.T) {
		h, err := ParseHash("383c9da631bd120169e82b0679e4c2e8d5050894")
		if !assert.NoError(t, err) {
			return
		}
		assert.False(t, h.Empty())
	})
	t.Run("returns new hash every time", func(t *testing.T) {
		h1 := EmptyHash()
		h2 := EmptyHash()
		assert.True(t, h1.Equals(h2))
		h1[0] = 10
		assert.False(t, h1.Equals(h2))
	})
}

func TestHash_Slice(t *testing.T) {
	t.Run("slice parsed hash", func(t *testing.T) {
		h, err := ParseHash("383c9da631bd120169e82b0679e4c2e8d5050894")
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, h, SliceToHash(h.Slice()))
	})
	t.Run("returns new slice every time", func(t *testing.T) {
		h, _ := ParseHash("383c9da631bd120169e82b0679e4c2e8d5050894")
		s1 := h.Slice()
		s2 := h.Slice()
		assert.Equal(t, s1, s2)
		s1[0] = 10
		assert.NotEqual(t, s1, s2)
	})
}

func TestParseHash(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		hash, err := ParseHash("383c9da631bd120169e82b0679e4c2e8d5050894")
		assert.NoError(t, err)
		assert.Equal(t, "383c9da631bd120169e82b0679e4c2e8d5050894", hash.String())
	})
	t.Run("ok - empty string input", func(t *testing.T) {
		hash, err := ParseHash("")
		assert.NoError(t, err)
		assert.True(t, hash.Empty())
	})
	t.Run("error - invalid input", func(t *testing.T) {
		hash, err := ParseHash("a23da")
		assert.Error(t, err)
		assert.True(t, hash.Empty())
	})
	t.Run("error - incorrect length", func(t *testing.T) {
		hash, err := ParseHash("383c9da631bd120169e82b0679e4c2e8d5050894383c9da631bd120169e82b0679e4c2e8d5050894")
		assert.EqualError(t, err, "incorrect hash length (40)")
		assert.True(t, hash.Empty())
	})
}

func TestHash_Equals(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		h1, _ := ParseHash("383c9da631bd120169e82b0679e4c2e8d5050894")
		h2, _ := ParseHash("383c9da631bd120169e82b0679e4c2e8d5050894")
		assert.True(t, h1.Equals(h2))
	})
	t.Run("not equal", func(t *testing.T) {
		h1 := EmptyHash()
		h2, _ := ParseHash("383c9da631bd120169e82b0679e4c2e8d5050894")
		assert.False(t, h1.Equals(h2))
	})
}

func TestHash_Compare(t *testing.T) {
	t.Run("smaller", func(t *testing.T) {
		h1, _ := ParseHash("0000000000000000000000000000000000000001")
		h2, _ := ParseHash("0000000000000000000000000000000000000002")
		assert.Equal(t, -1, h1.Compare(h2))
	})
	t.Run("equal", func(t *testing.T) {
		h1, _ := ParseHash("0000000000000000000000000000000000000001")
		h2, _ := ParseHash("0000000000000000000000000000000000000001")
		assert.Equal(t, 0, h1.Compare(h2))
	})
	t.Run("larger", func(t *testing.T) {
		h1, _ := ParseHash("0000000000000000000000000000000000000002")
		h2, _ := ParseHash("0000000000000000000000000000000000000001")
		assert.Equal(t, 1, h1.Compare(h2))
	})

}
