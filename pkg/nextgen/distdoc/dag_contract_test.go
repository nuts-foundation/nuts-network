package distdoc

import (
	"encoding/binary"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

// trackingVisitor just keeps track of which nodes were visited in what order.
type trackingVisitor struct {
	documents []Document
}

func (n *trackingVisitor) Accept(document Document) {
	n.documents = append(n.documents, document)
}

func (n trackingVisitor) JoinRefsAsString() string {
	var contents []string
	for _, document := range n.documents {
		val := strings.TrimLeft(model.Hash(document.Payload()).String(), "0")
		if val == "" {
			val = "0"
		}
		contents = append(contents, val)
	}
	return strings.Join(contents, ", ")
}

func DAGTest_Add(creator func(t *testing.T) DAG, t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		graph := creator(t)
		doc := testDocument(0)

		err := graph.Add(doc)

		assert.NoError(t, err)
		visitor := trackingVisitor{}
		err = graph.Walk(visitor.Accept)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, visitor.documents, 1)
		assert.Equal(t, doc.Ref(), visitor.documents[0].Ref())
	})
	t.Run("ok - out of order", func(t *testing.T) {
		_, documents := graphF(creator, t)
		graph := creator(t)

		for i := len(documents) - 1; i >= 0; i-- {
			err := graph.Add(documents[i])
			if !assert.NoError(t, err) {
				return
			}
		}

		visitor := trackingVisitor{}
		err := graph.Walk(visitor.Accept)
		if !assert.NoError(t, err) {
			return
		}
		assert.Regexp(t, "0, (1, 2|2, 1), (3, 4|4, 3), 5", visitor.JoinRefsAsString())
	})
	t.Run("error - cyclic graph", func(t *testing.T) {
		t.Skip("Algorithm for detecting cycles is not yet decided on")
		// A -> B -> C -> B
		A := testDocument(0)
		B := testDocument(1, A.Ref()).(*document)
		C := testDocument(2, B.Ref())
		B.prevs = append(B.prevs, C.Ref())

		graph := creator(t)
		err := graph.Add(A, B, C)
		assert.EqualError(t, err, "")
	})
}

func DAGTest_Walk(creator func(t *testing.T) DAG, t *testing.T) {
	t.Run("ok - walk graph F", func(t *testing.T) {
		visitor := trackingVisitor{}
		graph, _ := graphF(creator, t)

		err := graph.Walk(visitor.Accept)
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
		graph, docs := graphF(creator, t)
		G := testDocument(6, docs[3].Ref(), docs[5].Ref())
		graph.Add(G)

		graph.Walk(visitor.Accept)

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
		_, docs := graphF(creator, t)
		graph := creator(t)
		graph.Add(docs[0], docs[1], docs[3], docs[4], docs[5])

		graph.Walk(visitor.Accept)

		assert.Equal(t, "0, 1", visitor.JoinRefsAsString())
	})

	t.Run("ok - empty graph", func(t *testing.T) {
		graph := creator(t)
		visitor := trackingVisitor{}

		err := graph.Walk(visitor.Accept)
		if !assert.NoError(t, err) {
			return
		}

		assert.Empty(t, visitor.documents)
	})

	t.Run("ok - document added twice", func(t *testing.T) {
		graph := creator(t)
		d := testDocument(0)
		graph.Add(d)
		graph.Add(d)
		visitor := trackingVisitor{}

		graph.Walk(visitor.Accept)

		assert.Len(t, visitor.documents, 1)
	})

	t.Run("error - second root document", func(t *testing.T) {
		graph := creator(t)
		d1 := testDocument(0)
		d2 := testDocument(1)
		_ = graph.Add(d1)

		err := graph.Add(d2)
		assert.Equal(t, errRootAlreadyExists, err)
		visitor := trackingVisitor{}

		graph.Walk(visitor.Accept)

		// Assert the second root document wasn't added
		assert.Len(t, visitor.documents, 1)
		assert.Equal(t, d1.Ref(), visitor.documents[0].Ref())
	})
}

func DAGTest_MissingDocuments(creator func(t *testing.T) DAG, t *testing.T) {
	A := testDocument(0)
	B := testDocument(1, A.Ref())
	C := testDocument(2, B.Ref())
	t.Run("no missing documents (empty graph)", func(t *testing.T) {
		graph := creator(t)
		assert.Empty(t, graph.MissingDocuments())
	})
	t.Run("no missing documents (non-empty graph)", func(t *testing.T) {
		graph := creator(t)
		graph.Add(A, B, C)
		assert.Empty(t, graph.MissingDocuments())
	})
	t.Run("missing documents (non-empty graph)", func(t *testing.T) {
		graph := creator(t)
		graph.Add(A, C)
		assert.Len(t, graph.MissingDocuments(), 1)
		// Now add missing document B and assert there are no more missing documents
		graph.Add(B)
		assert.Empty(t, graph.MissingDocuments())
	})
}

// graphF creates the following graph:
//..................A
//................/  \
//...............B    C
//...............\   / \
//.................D    E
//.......................\
//........................F
func graphF(creator func(t *testing.T) DAG, t *testing.T) (DAG, []Document) {
	graph := creator(t)
	A := testDocument(0)
	B := testDocument(1, A.Ref())
	C := testDocument(2, A.Ref())
	D := testDocument(3, B.Ref(), C.Ref())
	E := testDocument(4, C.Ref())
	F := testDocument(5, E.Ref())
	docs := []Document{A, B, C, D, E, F}
	graph.Add(docs...)
	return graph, docs
}

var privateKey, certificate = generateKeyAndCertificate()

func testDocument(num uint32, prevs ...model.Hash) Document {
	ref := model.EmptyHash()
	binary.BigEndian.PutUint32(ref[model.HashSize-4:], num)
	unsignedDocument, _ := NewDocument(PayloadHash(ref), "foo/bar", prevs)
	signedDocument, _ := unsignedDocument.Sign(privateKey, certificate, time.Now())
	return signedDocument
}
