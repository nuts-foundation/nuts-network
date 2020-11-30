package distdoc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBFSWalker(t *testing.T) {
	t.Run("ok - walk graph F", func(t *testing.T) {
		visitor := trackingVisitor{}
		graph, _ := graphF(memoryDAGCreator, t)

		root, _ := graph.Root()
		err := graph.Walk(&BFSWalker{}, visitor.Accept, root)
		if !assert.NoError(t, err) {
			return
		}

		assert.Regexp(t, "0, (1, 2|2, 1), (3, 4|4, 3), 5", visitor.JoinRefsAsString())
	})

	t.Run("ok - walk graph G", func(t *testing.T) {
		//..................A
		//................/  \
		//...............B    C
		//...............\   / \
		//.................D    E
		//.................\.....\
		//..................\.....F
		//...................\.../
		//.....................G
		visitor := trackingVisitor{}
		graph, docs := graphF(memoryDAGCreator, t)
		G := testDocument(6, docs[3].Ref(), docs[5].Ref())
		graph.Add(G)

		root, _ := graph.Root()
		graph.Walk(&BFSWalker{}, visitor.Accept, root)

		assert.Regexp(t, "0, (1, 2|2, 1), (3, 4|4, 3), 5, 6", visitor.JoinRefsAsString())
	})

	t.Run("ok - walk graph F, C is missing", func(t *testing.T) {
		//..................A
		//................/  \
		//...............B    C (missing)
		//...............\   / \
		//.................D    E
		//.......................\
		//........................F
		visitor := trackingVisitor{}
		_, docs := graphF(memoryDAGCreator, t)
		graph := memoryDAGCreator(t)
		graph.Add(docs[0], docs[1], docs[3], docs[4], docs[5])

		root, _ := graph.Root()
		graph.Walk(&BFSWalker{}, visitor.Accept, root)

		assert.Equal(t, "0, 1", visitor.JoinRefsAsString())
	})

	t.Run("ok - empty graph", func(t *testing.T) {
		graph := memoryDAGCreator(t)
		visitor := trackingVisitor{}

		root, _ := graph.Root()
		err := graph.Walk(&BFSWalker{}, visitor.Accept, root)
		if !assert.NoError(t, err) {
			return
		}

		assert.Empty(t, visitor.documents)
	})

	t.Run("ok - document added twice", func(t *testing.T) {
		graph := memoryDAGCreator(t)
		d := testDocument(0)
		graph.Add(d)
		graph.Add(d)
		visitor := trackingVisitor{}

		root, _ := graph.Root()
		graph.Walk(&BFSWalker{}, visitor.Accept, root)

		assert.Len(t, visitor.documents, 1)
	})

	t.Run("error - second root document", func(t *testing.T) {
		graph := memoryDAGCreator(t)
		d1 := testDocument(0)
		d2 := testDocument(1)
		err := graph.Add(d1)

		err = graph.Add(d2)
		assert.Equal(t, errRootAlreadyExists, err)
		visitor := trackingVisitor{}

		root, _ := graph.Root()
		graph.Walk(&BFSWalker{}, visitor.Accept, root)

		assert.Len(t, visitor.documents, 1)
		assert.Equal(t, d1.Data(), visitor.documents[0].Data())
	})
}
