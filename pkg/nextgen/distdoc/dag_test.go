package distdoc

import (
	"crypto"
	"crypto/x509"
	"encoding/binary"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
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
		val := strings.TrimLeft(document.Ref().String(), "0")
		if val == "" {
			val = "0"
		}
		contents = append(contents, val)
	}
	return strings.Join(contents, ", ")
}

func TestChain_Add(t *testing.T) {
	key, certificate := generateKeyAndCertificate()
	t.Run("ok", func(t *testing.T) {
		graph := NewDAG()
		doc := testDocument(0)

		err := graph.Add(doc)

		assert.NoError(t, err)
		visitor := trackingVisitor{}
		graph.Walk(visitor.Accept)
		assert.Len(t, visitor.documents, 1)
		assert.Equal(t, doc.Ref(), visitor.documents[0].Ref())
	})
	t.Run("ok - out of order", func(t *testing.T) {
		_, documents := graphF(key, certificate)
		graph := NewDAG()

		for i := len(documents) - 1; i >= 0; i-- {
			graph.Add(documents[i])
		}

		visitor := trackingVisitor{}
		graph.Walk(visitor.Accept)
		assert.Equal(t, "0, 1, 2, 3, 4, 5", visitor.JoinRefsAsString())
	})
	t.Run("error - cyclic graph", func(t *testing.T) {
		t.Skip("Algorithm for detecting cycles is not yet decided on")
		// A -> B -> C -> B
		A := testDocument(0)
		B := testDocument(1, A.Ref()).(*document)
		C := testDocument(2, B.Ref())
		B.prevs = append(B.prevs, C.Ref())

		graph := NewDAG()
		err := graph.Add(A, B, C)
		assert.EqualError(t, err, "")
	})
}

func TestChain_Walk(t *testing.T) {
	key, certificate := generateKeyAndCertificate()

	t.Run("ok - walk graph F", func(t *testing.T) {
		visitor := trackingVisitor{}
		graph, _ := graphF(key, certificate)

		graph.Walk(visitor.Accept)

		assert.Equal(t, "0, 1, 2, 3, 4, 5", visitor.JoinRefsAsString())
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
		graph, docs := graphF(key, certificate)
		G := testDocument(6, docs[3].Ref(), docs[5].Ref())
		graph.Add(G)

		graph.Walk(visitor.Accept)

		assert.Equal(t, "0, 1, 2, 3, 4, 5, 6", visitor.JoinRefsAsString())
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
		_, docs := graphF(key, certificate)
		graph := NewDAG()
		graph.Add(docs[0], docs[1], docs[3], docs[4], docs[5])

		graph.Walk(visitor.Accept)

		assert.Equal(t, "0, 1", visitor.JoinRefsAsString())
	})

	t.Run("ok - empty graph", func(t *testing.T) {
		graph := NewDAG()
		visitor := trackingVisitor{}

		graph.Walk(visitor.Accept)

		assert.Empty(t, visitor.documents)
	})

	t.Run("ok - document added twice", func(t *testing.T) {
		graph := NewDAG()
		d := testDocument(0)
		graph.Add(d)
		err := graph.Add(d)
		assert.NoError(t, err)
		visitor := trackingVisitor{}

		graph.Walk(visitor.Accept)

		assert.Len(t, visitor.documents, 1)
	})

	t.Run("error - second root document", func(t *testing.T) {
		graph := NewDAG()
		d1 := testDocument(0 )
		d2 := testDocument(1)
		graph.Add(d1)
		err := graph.Add(d2)
		assert.Equal(t, errRootAlreadyExists, err)
		visitor := trackingVisitor{}

		graph.Walk(visitor.Accept)

		assert.Len(t, visitor.documents, 1)
		assert.Equal(t, d1, visitor.documents[0])
	})
}


func TestDAG_MissingDocuments(t *testing.T) {
	A := testDocument(0)
	B := testDocument(1, A.Ref())
	C := testDocument(2, B.Ref())
	t.Run("no missing documents (empty graph)", func(t *testing.T) {
		graph := NewDAG()
		assert.Empty(t, graph.MissingDocuments())
	})
	t.Run("no missing documents (non-empty graph)", func(t *testing.T) {
		graph := NewDAG()
		graph.Add(A, B, C)
		assert.Empty(t, graph.MissingDocuments())
	})
	t.Run("missing documents (non-empty graph)", func(t *testing.T) {
		graph := NewDAG()
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
func graphF(key crypto.Signer, certificate *x509.Certificate) (*DAG, []Document) {
	graph := NewDAG()
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

func testDocument(num uint32, prevs ...model.Hash) Document {
	ref := model.EmptyHash()
	binary.BigEndian.PutUint32(ref[model.HashSize-4:], num)
	return &document{
		prevs:       prevs,
		payload:     ref,
		payloadType: "foo/bar",
		ref:         ref,
	}
}
