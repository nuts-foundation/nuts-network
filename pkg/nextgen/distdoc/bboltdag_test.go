package distdoc

import (
	"crypto/sha1"
	"github.com/nuts-foundation/nuts-go-test/io"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/stretchr/testify/assert"
	"path"
	"testing"
)

var bboltDAGCreator = func(t *testing.T) DAG {
	if dag, err := NewBBoltDAG(path.Join(io.TestDirectory(t), "dag.db")); err != nil {
		t.Fatal(err)
		return nil
	} else {
		return dag
	}
}

func TestBBoltDAG_Add(t *testing.T) {
	DAGTest_Add(bboltDAGCreator, t)
}

func TestBBoltDAG_MissingDocuments(t *testing.T) {
	DAGTest_MissingDocuments(bboltDAGCreator, t)
}

func TestBBoltDAG_Walk(t *testing.T) {
	DAGTest_Walk(bboltDAGCreator, t)
}

func Test_parseHashList(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		assert.Empty(t, parseHashList([]byte{}))
	})
	t.Run("1 entry", func(t *testing.T) {
		h1 := sha1.Sum([]byte("Hello, World!"))
		actual := parseHashList(h1[:])
		assert.Len(t, actual, 1)
		assert.Equal(t, model.SliceToHash(h1[:]), actual[0])
	})
	t.Run("2 entries", func(t *testing.T) {
		h1 := sha1.Sum([]byte("Hello, World!"))
		h2 := sha1.Sum([]byte("Hello, All!"))
		actual := parseHashList(append(h1[:], h2[:]...))
		assert.Len(t, actual, 2)
		assert.Equal(t, model.SliceToHash(h1[:]), actual[0])
		assert.Equal(t, model.SliceToHash(h2[:]), actual[1])
	})
	t.Run("2 entries, dangling bytes", func(t *testing.T) {
		h1 := sha1.Sum([]byte("Hello, World!"))
		h2 := sha1.Sum([]byte("Hello, All!"))
		input := append(h1[:], h2[:]...)
		input = append(input, 1, 2, 3) // Add some dangling bytes
		actual := parseHashList(input)
		assert.Len(t, actual, 2)
		assert.Equal(t, model.SliceToHash(h1[:]), actual[0])
		assert.Equal(t, model.SliceToHash(h2[:]), actual[1])
	})
}