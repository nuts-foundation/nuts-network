package distdoc

import (
	"container/list"
	"errors"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"sort"
)

var errRootAlreadyExists = errors.New("root document already exists")

// DAG is a directed acyclic graph consisting of nodes (documents) referring to preceding nodes.
type DAG interface {
	// Add adds one or more documents to the DAG. If it can't be added an error is returned. Nil entries are ignored.
	Add(documents ...Document) error
	// MissingDocuments returns the hashes of the documents we know we are missing and should still be resolved.
	MissingDocuments() []model.Hash
	// Walk visits every node of the DAG, starting at the given hash working its way down each level until every leaf is visited.
	// when startAt is an empty hash, the walker starts at the root node.
	Walk(walker Walker, visitor Visitor, startAt model.Hash) error
	// Root returns the root hash of the DAG. If there's no root an empty hash is returned. If an error occurs, it is returned.
	Root() (model.Hash, error)
}

type Walker interface {
	// walk visits every node of the DAG, starting at the given start node and working down each level until every leaf is visited.
	// numberOfNodes is an indicative number of nodes that's expected to be visited. It's used for optimizing memory usage.
	// getFn is a function for reading a document from the DAG using the given ref hash. If not found nil must be returned.
	// nextsFn is a function for reading a document's nexts using the given ref hash. If not found nil must be returned.
	walk(visitor Visitor, startAt model.Hash, getFn func(model.Hash) (Document, error), nextsFn func(model.Hash) ([]model.Hash, error), numberOfNodes int) error
}

// BFSWalker walks the DAG using the Breadth-First-Search (BFS) as described by Anany Levitin in "The Design & Analysis of Algorithms".
// It visits the whole tree level for level (breadth first vs depth first). It works by taking a node from queue and
// then adds the node's children (downward edges) to the queue. It starts by adding the root node to the queue and
// loops over the queue until empty, meaning all nodes reachable from the root node have been visited. Since our
// DAG contains merges (two parents referring to the same child node) we also keep a map to avoid visiting a
// merger node twice.
//
// This also means we have to make sure we don't visit the merger node before all of its previous nodes have been
// visited, which BFS doesn't account for. If that happens we skip the node without marking it as visited,
// so it will be visited again when the unvisited previous node is visited, which re-adds the merger node to the queue.
type BFSWalker struct{}

func (w BFSWalker) walk(visitor Visitor, startAt model.Hash, getFn func(model.Hash) (Document, error), nextsFn func(model.Hash) ([]model.Hash, error), numberOfNodes int) error {
	queue := list.New()
	queue.PushBack(startAt)
	visitedDocuments := make(map[model.Hash]bool, numberOfNodes)

ProcessQueueLoop:
	for queue.Len() > 0 {
		// Pop first element of queue
		front := queue.Front()
		queue.Remove(front)
		currentRef := front.Value.(model.Hash)

		// Make sure we haven't already visited this node
		if _, visited := visitedDocuments[currentRef]; visited {
			continue
		}

		// Make sure all prevs have been visited. Otherwise just continue, it will be re-added to the queue when the
		// unvisited prev node is visited and re-adds this node to the processing queue.
		currentDocument, err := getFn(currentRef)
		if err != nil {
			return err
		}

		for _, prev := range currentDocument.Previous() {
			if _, visited := visitedDocuments[prev]; !visited {
				continue ProcessQueueLoop
			}
		}

		// Add child nodes to processing queue
		// Processing order of nodes on the same level doesn't really matter for correctness of the DAG travel
		// but it makes testing easier.
		if nexts, err := nextsFn(currentRef); err != nil {
			return err
		} else if nexts != nil {
			sortedEdges := make([]model.Hash, 0, len(nexts))
			for _, nextNode := range nexts {
				sortedEdges = append(sortedEdges, nextNode)
			}
			sort.Slice(sortedEdges, func(i, j int) bool {
				return sortedEdges[i].Compare(sortedEdges[j]) < 0
			})
			for _, nextNode := range sortedEdges {
				queue.PushBack(nextNode)
			}
		}

		// Visit the node
		visitor(currentDocument)

		// Mark this node as visited
		visitedDocuments[currentRef] = true
	}
	return nil
}

// Visitor defines the contract for a function that visits the DAG.
type Visitor func(document Document)
