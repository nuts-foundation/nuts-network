package distdoc

import (
	testIO "github.com/nuts-foundation/nuts-go-test/io"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemoryEliasDBDAG_Add(t *testing.T) {
	DAGTest_Add(inMemoryEliasDBDAGCreator, t)
}

func TestMemoryEliasDBDAG_MissingDocuments(t *testing.T) {
	DAGTest_MissingDocuments(inMemoryEliasDBDAGCreator, t)
}

func TestMemoryEliasDBDAG_Walk(t *testing.T) {
	DAGTest_Walk(inMemoryEliasDBDAGCreator, t)
}

func TestDiskEliasDBDAG_Add(t *testing.T) {
	DAGTest_Add(onDiskEliasDBDAGCreator, t)
}

func TestDiskEliasDBDAG_MissingDocuments(t *testing.T) {
	DAGTest_MissingDocuments(onDiskEliasDBDAGCreator, t)
}

func TestDiskEliasDBDAG_Walk(t *testing.T) {
	DAGTest_Walk(onDiskEliasDBDAGCreator, t)
}

func TestEliasDBDAG_StoreAndLoad(t *testing.T) {
	dag := onDiskEliasDBDAGCreator(t)
	// We need some variety to test persistence of array of previous document refs
	doc1 := testDocument(1)
	doc2 := testDocument(2, doc1.Ref())
	doc3 := testDocument(3, doc1.Ref())
	doc4 := testDocument(4, doc2.Ref(), doc3.Ref())
	err := dag.Add(doc1, doc2, doc3, doc4)
	if !assert.NoError(t, err) {
		return
	}
	actual := make([]Document, 0)
	err = dag.Walk(func(document Document) {
		actual = append(actual, document)
	})
	if !assert.NoError(t, err) {
		return
	}
	assert.Len(t, actual, 4)
	assert.Equal(t, doc1, actual[0])
	// 1 and 2 can appear in any order
	if !(assert.ObjectsAreEqual(doc2, actual[1]) && assert.ObjectsAreEqual(doc3, actual[2])) {
		if !(assert.ObjectsAreEqual(doc2, actual[2]) && assert.ObjectsAreEqual(doc3, actual[1])) {
			t.Fatal("doc 1 and 2 not in correct order/not equal")
		}
	}
	assert.Equal(t, doc4, actual[3])
}

var eliasDBDAGCreator = func(creator func() (DAG, error), t *testing.T) DAG {
	dag, err := creator()
	if err != nil {
		t.Fatal("Failed to create EliasDB DAG", err)
	}
	// Closing the DAG causes a race condition, which fails the build. Bug is reported to library author:
	// https://github.com/krotik/eliasdb/issues/27
	// Should be enabled again after it's fixed.
	//t.Cleanup(func() {
	//	if closer, ok := dag.(io.Closer); ok {
	//		if err := closer.Close(); err != nil {
	//			logrus.Error("Unable to close DAG:", err)
	//		}
	//	}
	//})
	return dag
}

var inMemoryEliasDBDAGCreator = func(t *testing.T) DAG {
	return eliasDBDAGCreator(func() (DAG, error) {
		return NewEliasDBDAG(NewMemoryEliasDB().Manager), nil
	}, t)
}

var onDiskEliasDBDAGCreator = func(t *testing.T) DAG {
	return eliasDBDAGCreator(func() (DAG, error) {
		if db, err := NewDiskEliasDB(testIO.TestDirectory(t)); err != nil {
			return nil, err
		} else {
			return NewEliasDBDAG(db.Manager), nil
		}
	}, t)
}
