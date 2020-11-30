package distdoc

import (
	"github.com/nuts-foundation/nuts-network/pkg/model"
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


func (c memoryDAG) Root() (model.Hash, error) {
	if c.root == nil {
		return model.EmptyHash(), nil
	} else {
		return c.root.document.Ref(), nil
	}
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

func (c memoryDAG) Walk(walker Walker, visitor Visitor, startAt model.Hash) error {
	if c.root == nil {
		return nil
	}
	return walker.walk(visitor, startAt, func(hash model.Hash) (Document, error) {
		if node := c.nodes[hash]; node == nil {
			return nil, nil
		} else {
			return node.document, nil
		}
	}, func(hash model.Hash) ([]model.Hash, error) {
		if node := c.nodes[hash]; node == nil {
			return nil, nil
		} else {
			nexts := make([]model.Hash, 0)
			for k, _ := range node.edges {
				nexts = append(nexts, k)
			}
			return nexts, nil
		}
	}, len(c.nodes))
}
