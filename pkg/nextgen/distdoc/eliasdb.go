package distdoc

import (
	"container/list"
	"crypto/x509"
	"devt.de/krotik/eliasdb/graph"
	"devt.de/krotik/eliasdb/graph/data"
	"devt.de/krotik/eliasdb/graph/graphstorage"
	"encoding/gob"
	"fmt"
	"github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"sort"
	"strings"
	"time"
)

const partition = "main"
const documentKind = "document"
const prevEdgeKind = "prev"

// resolvedAttr is a bool attribute indicating we have the document. If false we have a succeeding document with a prev
// pointing to this document we still have to resolve.
const resolvedAttr = "resolved"

// rootAttr indicates the document is the DAGs root document. There can only be one document with this attribute being true.
const rootAttr = "root"

const (
	// documentAttrPrefix is the prefix for document properties when mapping them to EliasDB node attributes
	documentAttrPrefix          = "document"
	documentPayloadTypeAttr     = "typ"
	documentPayloadAttr         = "payload"
	documentPreviousAttr        = "prev"
	documentVersionAttr         = "ver"
	documentTimelineIDAttr      = "tid"
	documentTimelineVersionAttr = "tiv"
	documentSigningTimeAttr     = "time"
	documentSigningCertAttr     = "certificate"
	documentDataAttr            = "data"
	documentPayloadDataAttr     = "payloadData"
)

const errInvalidDocumentNode = "invalid document node in EliasDB: %s: %v"
const errDocumentNodeNotFound = "document node in EliasDB not found: %s"

func init() {
	// Register custom gob types
	gob.Register(PayloadHash{})
	gob.Register(model.Hash{})
	gob.Register([]model.Hash{})
	gob.Register(time.Time{})
	gob.Register(Version(0))
}

type eliasDBDAG struct {
	storage    graphstorage.Storage
	manager    *graph.Manager
	projectors []Projector
}

func (e *eliasDBDAG) UpdateProjections(documentRef model.Hash) error {
	node, err := e.manager.FetchNode(partition, documentRef.String(), documentKind)
	if err != nil {
		return err
	} else if node == nil {
		return fmt.Errorf(errDocumentNodeNotFound, documentRef)
	}
	document, err := nodeToDocument(node)
	if err != nil {
		return err
	}
	if err := e.applyProjectors(document, node); err != nil {
		return fmt.Errorf("error while applying projectors for document %s: %w", documentRef, err)
	}
	if err := e.manager.StoreNode(partition, node); err != nil {
		return fmt.Errorf("error while storing updated node %s: %v", documentRef, err)
	}
	return nil
}

func (e *eliasDBDAG) RegisterProjector(projector Projector) {
	e.projectors = append(e.projectors, projector)
}

func (e eliasDBDAG) MissingDocuments() []model.Hash {
	query, err := e.manager.NodeIndexQuery(partition, documentKind)
	if err != nil {
		log.Log().Error("Unable to query missing documents", err)
		return nil
	} else if query == nil {
		return nil
	}
	result := make([]model.Hash, 0)
	if refs, err := query.LookupValue(resolvedAttr, "false"); err != nil {
		log.Log().Error("Unable to query missing documents", err)
		return nil
	} else {
		for _, ref := range refs {
			if hash, err := model.ParseHash(ref); err != nil {
				log.Log().Errorf("Invalid ref %s: %v", ref, err)
			} else {
				result = append(result, hash)
			}
		}
	}
	return result
}

func (e eliasDBDAG) Add(documents ...Document) error {
	for _, document := range documents {
		if err := e.add(document); err != nil {
			return err
		}
	}
	return nil
}

func (e eliasDBDAG) add(document Document) error {
	transaction := graph.NewGraphTrans(e.manager)
	node := e.createNodeFromDocument(document)
	node.SetAttr(resolvedAttr, true)
	// If root document, assert we don't already have one
	if len(document.Previous()) == 0 {
		node.SetAttr(rootAttr, true)
		existingRootNode, err := e.lookupRootNode()
		if err != nil {
			return fmt.Errorf("error while checking for existing root node: %w", err)
		} else if existingRootNode != nil {
			return errRootAlreadyExists
		}
	}
	// Store node
	if err := transaction.StoreNode(partition, node); err != nil {
		return fmt.Errorf("unable to store node: %w", err)
	}
	// Store edges to previous documents
	for _, prev := range document.Previous() {
		// TODO: Make this existence check smarter
		prevNode, err := e.manager.FetchNode(partition, prev.String(), documentKind)
		if err != nil {
			return fmt.Errorf("unable to fetch prev %s: %w", prev, err)
		} else if prevNode == nil {
			prevNode = e.createNode(prev)
			prevNode.SetAttr(resolvedAttr, false)
			if err := transaction.StoreNode(partition, prevNode); err != nil {
				return fmt.Errorf("unable to store prev node: %w", err)
			}
		}
		edge := data.NewGraphEdge()
		edge.SetAttr(data.NodeKey, prevNode.Key()+"->"+node.Key())
		edge.SetAttr(data.NodeKind, prevEdgeKind)
		edge.SetAttr(data.EdgeEnd1Key, prevNode.Key())
		edge.SetAttr(data.EdgeEnd1Kind, prevNode.Kind())
		edge.SetAttr(data.EdgeEnd1Role, "left")
		edge.SetAttr(data.EdgeEnd1Cascading, false) // TODO: check this
		edge.SetAttr(data.EdgeEnd2Key, node.Key())
		edge.SetAttr(data.EdgeEnd2Kind, node.Kind())
		edge.SetAttr(data.EdgeEnd2Role, "right")
		edge.SetAttr(data.EdgeEnd2Cascading, false) // TODO: check this
		if err = transaction.StoreEdge(partition, edge); err != nil {
			return fmt.Errorf("unable to store edge to prev %s: %w", prev, err)
		}
	}
	return transaction.Commit()
}

func (e eliasDBDAG) lookupRootNode() (data.Node, error) {
	query, err := e.manager.NodeIndexQuery(partition, documentKind)
	if err != nil {
		return nil, err
	} else if query == nil {
		// Graph is empty
		return nil, nil
	}
	rootKeys, err := query.LookupValue(rootAttr, "true")
	if err != nil {
		return nil, err
	}
	if len(rootKeys) > 1 {
		return nil, fmt.Errorf("multiple root documents found: %v", rootKeys)
	} else if len(rootKeys) == 0 {
		return nil, nil
	}
	node, err := e.manager.FetchNode(partition, rootKeys[0], documentKind)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch root key %s: %w", rootKeys[0], err)
	}
	return node, nil
}

func (e eliasDBDAG) createNode(ref model.Hash) data.Node {
	node := data.NewGraphNode()
	node.SetAttr(data.NodeKey, ref.String())
	node.SetAttr(data.NodeKind, documentKind)
	return node
}

func (e *eliasDBDAG) createNodeFromDocument(document Document) data.Node {
	node := e.createNode(document.Ref())
	_ = e.applyProjector(document, node, &documentProjector{})
	return node
}

func (e eliasDBDAG) applyProjectors(document Document, node data.Node) error {
	for _, projector := range e.projectors {
		if err := e.applyProjector(document, node, projector); err != nil {
			return err
		}
	}
	return nil
}

func (e eliasDBDAG) applyProjector(document Document, node data.Node, projector Projector) error {
	projection, err := projector.Project(&e, document)
	if err != nil {
		return fmt.Errorf("projector returned an error: %v", err)
	}
	if projection == nil {
		// This projection didn't project anything, which is fine.
		return nil
	}
	if strings.TrimSpace(projection.Name) == "" {
		return fmt.Errorf("projection has invalid name: %s", projection.Name)
	}
	for key, value := range projection.Properties {
		node.SetAttr(fmt.Sprintf("%s.%s", projection.Name, key), value)
	}
	return nil
}

func (e eliasDBDAG) Walk(visitor Visitor) error {
	rootNode, err := e.lookupRootNode()
	if err != nil {
		return err
	} else if rootNode == nil {
		// Empty graph
		return nil
	}
	rootDocument, err := nodeToDocument(rootNode)
	if err != nil {
		return fmt.Errorf(errInvalidDocumentNode, rootNode.Key(), err)
	}
	// Copy of breadth-first-search algorithm from memory DAG
	queue := list.New()
	queue.PushBack(rootDocument)
	visitedDocuments := make(map[string]bool, e.manager.NodeCount(documentKind)) // TODO: maybe make this smarter
ProcessQueueLoop:
	for queue.Len() > 0 {
		// Pop first element of queue
		front := queue.Front()
		queue.Remove(front)
		currentDocument := front.Value.(Document)
		currentDocumentRef := currentDocument.Ref().String()

		// Make sure we haven't already visited this node
		if _, visited := visitedDocuments[currentDocumentRef]; visited {
			continue
		}

		// Make sure all prevs have been visited. Otherwise just continue, it will be re-added to the queue when the
		// unvisited prev node is visited and re-adds this node to the processing queue.
		for _, prev := range currentDocument.Previous() {
			if _, visited := visitedDocuments[prev.String()]; !visited {
				continue ProcessQueueLoop
			}
		}

		// Add child nodes to processing queue
		// Processing order of nodes on the same level doesn't really matter for correctness of the DAG travel
		// but it makes testing easier.
		nextNodes, _, err := e.manager.Traverse(partition, currentDocumentRef, documentKind, "left:"+prevEdgeKind+":right:"+documentKind, true)
		if err != nil {
			return err
		}
		if len(nextNodes) > 0 {
			sortedNextDocuments := make([]Document, 0, len(nextNodes))
			for _, nextNode := range nextNodes {
				if nextDocument, err := nodeToDocument(nextNode); err != nil {
					return fmt.Errorf(errInvalidDocumentNode, nextNode.Key(), err)
				} else {
					sortedNextDocuments = append(sortedNextDocuments, nextDocument)
				}
			}
			sortDocumentsOnHash(sortedNextDocuments)
			for _, nextDocument := range sortedNextDocuments {
				queue.PushBack(nextDocument)
			}
		}

		// Visit the node
		visitor(currentDocument)

		// Mark this node as visited
		visitedDocuments[currentDocumentRef] = true
	}
	return nil
}

func sortDocumentsOnHash(slice []Document) {
	sort.Slice(slice, func(i, j int) bool {
		return slice[i].Ref().Compare(slice[j].Ref()) < 0
	})
}

type EliasDB struct {
	Storage graphstorage.Storage
	Manager *graph.Manager
}

func NewDiskEliasDB(directory string) (*EliasDB, error) {
	if storage, err := graphstorage.NewDiskGraphStorage(directory, false); err != nil {
		return nil, err
	} else {
		return &EliasDB{Storage: storage, Manager: graph.NewGraphManager(storage)}, nil
	}
}

func NewMemoryEliasDB() *EliasDB {
	storage := graphstorage.NewMemoryGraphStorage("db")
	return &EliasDB{Storage: storage, Manager: graph.NewGraphManager(storage)}
}

func (e EliasDB) Close() error {
	return e.Storage.Close()
}

func NewEliasDBDAG(manager *graph.Manager) ProjectableDAG {
	return &eliasDBDAG{
		manager:    manager,
		projectors: make([]Projector, 0),
	}
}

type eliasDBPayloadStore struct {
	manager *graph.Manager
}

func (e eliasDBPayloadStore) WritePayload(documentRef model.Hash, data []byte) error {
	node, err := e.manager.FetchNode(partition, documentRef.String(), documentKind)
	if err != nil {
		return fmt.Errorf("unable to fetch node %s for storing payload: %w", documentRef, err)
	}
	if node == nil {
		return fmt.Errorf(errDocumentNodeNotFound, documentRef)
	}
	node.SetAttr(documentPayloadDataAttr, data)
	return e.manager.StoreNode(partition, node)
}

func (e eliasDBPayloadStore) ReadPayload(documentRef model.Hash) ([]byte, error) {
	node, err := e.manager.FetchNodePart(partition, documentRef.String(), documentKind, []string{documentPayloadDataAttr})
	if err != nil {
		return nil, fmt.Errorf("unable to read payload of %s: %w", documentRef, err)
	}
	if node == nil {
		return nil, fmt.Errorf(errDocumentNodeNotFound, documentRef)
	}
	if data, ok := node.Attr(documentPayloadDataAttr).([]byte); !ok {
		return nil, nil
	} else {
		return data, nil
	}
}

func NewEliasDBPayloadStore(manager *graph.Manager) PayloadStore {
	return &eliasDBPayloadStore{manager: manager}
}

func nodeToDocument(node data.Node) (Document, error) {
	result := document{}

	invalidValue := func(field string) error {
		return fmt.Errorf("invalid EliasDB document field: %s", field)
	}
	invalidValueWithErr := func(field string, err error) error {
		return fmt.Errorf("invalid EliasDB document field: %s: %v", field, err)
	}
	attr := func(name string) interface{} {
		return node.Attr(documentAttrPrefix + "." + name)
	}
	var ok bool
	var err error

	if result.payloadType, ok = attr(documentPayloadTypeAttr).(string); !ok {
		return nil, invalidValue(documentPayloadTypeAttr)
	}
	if result.data, ok = attr(documentDataAttr).([]byte); !ok {
		return nil, invalidValue(documentDataAttr)
	}
	if result.payload, ok = attr(documentPayloadAttr).(model.Hash); !ok {
		return nil, invalidValue(documentPayloadAttr)
	}
	if result.prevs, ok = attr(documentPreviousAttr).([]model.Hash); !ok {
		return nil, invalidValue(documentPreviousAttr)
	}
	if result.ref, err = model.ParseHash(node.Attr(data.NodeKey).(string)); err != nil {
		return nil, invalidValue(data.NodeKey)
	}
	if result.version, ok = attr(documentVersionAttr).(Version); !ok {
		return nil, invalidValue(documentVersionAttr)
	}
	if result.timelineID, ok = attr(documentTimelineIDAttr).(model.Hash); !ok {
		return nil, invalidValue(documentTimelineIDAttr)
	}
	if result.timelineVersion, ok = attr(documentTimelineVersionAttr).(int); !ok {
		return nil, invalidValue(documentTimelineVersionAttr)
	}
	if signingCertAsBytes, ok := attr(documentSigningCertAttr).([]byte); !ok {
		return nil, invalidValue(documentSigningCertAttr)
	} else if result.certificate, err = x509.ParseCertificate(signingCertAsBytes); err != nil {
		return nil, invalidValueWithErr(documentSigningCertAttr, err)
	}
	if result.signingTime, ok = attr(documentSigningTimeAttr).(time.Time); !ok {
		return nil, invalidValue(documentSigningTimeAttr)
	}

	return &result, nil
}

type documentProjector struct{}

func (_ documentProjector) Project(dag ProjectableDAG, document Document) (*Projection, error) {
	props := map[string]interface{}{
		documentPayloadTypeAttr:     document.PayloadType(),
		documentPayloadAttr:         model.Hash(document.Payload()),
		documentPreviousAttr:        document.Previous(),
		documentVersionAttr:         document.Version(),
		documentTimelineIDAttr:      document.TimelineID(),
		documentTimelineVersionAttr: document.TimelineVersion(),
		documentSigningTimeAttr:     document.SigningTime(),
		documentSigningCertAttr:     document.SigningCertificate().Raw,
		documentDataAttr:            document.Data(),
	}
	return &Projection{
		Name:       documentAttrPrefix,
		Properties: props,
	}, nil
}
