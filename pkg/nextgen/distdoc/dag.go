package distdoc

import (
	"errors"
	"github.com/nuts-foundation/nuts-network/pkg/model"
)

var errRootAlreadyExists = errors.New("root document already exists")

// DAG is a directed acyclic graph consisting of nodes (documents) referring to preceding nodes.
type DAG interface {
	// Add adds one or more documents to the DAG. If it can't be added an error is returned. Nil entries are ignored.
	Add(documents ...Document) error
	// Walk visits every node of the DAG, starting at the root node and working down each level until every leaf is visited.
	Walk(visitor Visitor) error
	// MissingDocuments returns the hashes of the documents we know we are missing and should still be resolved.
	MissingDocuments() []model.Hash
}

type ProjectableDAG interface {
	DAG
	// RegisterProjector registers a projection which is called when a document is added.
	RegisterProjector(projector Projector)
	// UpdateProjections invokes the projectors for the given document. This could be used to update projections when
	// external factors changed which influence the projection (e.g. previously missing document contents were received).
	UpdateProjections(documentRef model.Hash) error
}

// Projector wraps the Project function which creates a property-based view (a.k.a. projection) of a document.
// The projection is stored with the document for querying.
type Projector interface {
	Project(dag ProjectableDAG, document Document) (*Projection, error)
}

type Projection struct {
	Name       string
	Properties map[string]interface{}
}

// Visitor defines the contract for a function that visits the DAG.
type Visitor func(document Document)
