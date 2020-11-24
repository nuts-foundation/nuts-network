package distdoc

import (
	"container/list"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"sort"
)

type memoryDAG struct {
	// root contains the root node of the DAG
	root  *node
	nodes map[model.Hash]*node
	// missingEdges contains the forward references (A -> [B, C]) which couldn't be updated since documents were added out-of-order (B and C were added before A).
	missingEdges map[model.Hash][]*node
}

type node struct {
	document Document
	// edges contains forward references between documents (A -> B) so we can walk the DAG top to bottom more easily.
	// This is the inverse of the 'prevs' field which a backreference (B -> A).
	edges map[model.Hash]*node
}

func NewMemoryDAG() DAG {
	return &memoryDAG{
		nodes:        map[model.Hash]*node{},
		missingEdges: map[model.Hash][]*node{},
	}
}

func (c memoryDAG) MissingDocuments() []model.Hash {
	result := make([]model.Hash, 0, len(c.missingEdges))
	for hash, _ := range c.missingEdges {
		result = append(result, hash.Clone())
	}
	return result
}

func (c *memoryDAG) Add(documents ...Document) error {
	for _, document := range documents {
		if document != nil {
			if err := c.add(document); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *memoryDAG) add(document Document) error {
	// TODO: Check cyclic -> error
	ref := document.Ref()
	if _, ok := c.nodes[ref]; ok {
		// We already have this node
		return nil
	}
	newEntry := node{
		document: document,
		edges:    map[model.Hash]*node{},
	}
	// Check if this is a root document
	if len(document.Previous()) == 0 {
		if c.root == nil {
			c.root = &newEntry
		} else {
			return errRootAlreadyExists
		}
	}
	// Make sure nodes' edges field is kept up to date with a pointer to the new document (A -> B)
	for _, prev := range document.Previous() {
		if prevEntry, ok := c.nodes[prev]; ok {
			prevEntry.edges[ref] = &newEntry
		} else {
			// We don't have this previous document yet, mark it as a missing edge so we can quickly update the A -> B
			// edge when we receive B. This is the case when documents are added out-of-order.
			missingEdges := c.missingEdges[prev]
			c.missingEdges[prev] = append(missingEdges, &newEntry)
		}
	}
	// Update forward references (A -> B) for documents that refer to this document but were added earlier (out-of-order)
	missingEdges := c.missingEdges[ref]
	if len(missingEdges) > 0 {
		for _, nextNode := range missingEdges {
			newEntry.edges[nextNode.document.Ref()] = nextNode
		}
		delete(c.missingEdges, ref)
	}
	c.nodes[ref] = &newEntry
	return nil
}

func (c memoryDAG) Walk(visitor Visitor) error {
	if c.root == nil {
		return nil
	}
	// The algorithm below is a Breadth-First-Search (BFS) as described by Anany Levitin in "The Design & Analysis of Algorithms".
	// It visits the whole tree level for level (breadth first vs depth first). It works by taking a node from queue and
	// then adds the node's children (downward edges) to the queue. It starts by adding the root node to the queue and
	// loops over the queue until empty, meaning all nodes reachable from the root node have been visited. Since our
	// DAG contains merges (two parents referring to the same child node) we also keep a map to avoid visiting a
	// merger node twice.
	//
	// This also means we have to make sure we don't visit the merger node before all of its previous nodes have been
	// visited, which BFS doesn't account for. If that happens we skip the node without marking it as visited,
	// so it will be visited again when the unvisited previous node is visited, which re-adds the merger node to the queue.
	queue := list.New()
	queue.PushBack(c.root)
	visitedNodes := make(map[*node]bool, len(c.nodes))
	visitedDocuments := make(map[model.Hash]bool, len(c.nodes))
ProcessQueueLoop:
	for queue.Len() > 0 {
		// Pop first element of queue
		front := queue.Front()
		queue.Remove(front)
		currentNode := front.Value.(*node)

		// Make sure we haven't already visited this node
		if _, visited := visitedNodes[currentNode]; visited {
			continue
		}

		// Make sure all prevs have been visited. Otherwise just continue, it will be re-added to the queue when the
		// unvisited prev node is visited and re-adds this node to the processing queue.
		for _, prev := range currentNode.document.Previous() {
			if _, visited := visitedDocuments[prev]; !visited {
				continue ProcessQueueLoop
			}
		}

		// Add child nodes to processing queue
		// Processing order of nodes on the same level doesn't really matter for correctness of the DAG travel
		// but it makes testing easier.
		if currentNode.edges != nil {
			sortedEdges := make([]*node, 0, len(currentNode.edges))
			for _, nextNode := range currentNode.edges {
				sortedEdges = append(sortedEdges, nextNode)
			}
			sort.Slice(sortedEdges, func(i, j int) bool {
				return sortedEdges[i].document.Ref().Compare(sortedEdges[j].document.Ref()) < 0
			})
			for _, nextNode := range sortedEdges {
				queue.PushBack(nextNode)
			}
		}

		// Visit the node
		visitor(currentNode.document)

		// Mark this node as visited
		visitedNodes[currentNode] = true
		visitedDocuments[currentNode.document.Ref()] = true
	}
	return nil
}
