package distdoc

import (
	"fmt"
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"go.etcd.io/bbolt"
)

// documentsBucket is the name of the Bolt bucket that holds the actual documents as JSON.
const documentsBucket = "documents"

// missingDocumentsBucket is the name of the Bolt bucket that holds the references of the documents we're having prevs
// to, but are missing (and will be added later, hopefully).
const missingDocumentsBucket = "missingdocuments"

// nextsBucket is the name of the Bolt bucket that holds the forward document references (a.k.a. "nexts") as document
// refs. The value ([]byte) should be split in chunks of HashSize where each entry is a forward reference (next).
const nextsBucket = "nexts"

// rootDocumentKey is the name of the bucket entry that holds the refs of the root documents.
const rootsDocumentKey = "roots"

// boltDBFileMode holds the Unix file mode the created BBolt database files will have.
const boltDBFileMode = 0600

type bboltDAG struct {
	db *bbolt.DB
}

// NewBBoltDAG creates a etcd/bbolt backed DAG using the given database file path. If the file doesn't exist, it's created.
// The parent directory of the path must exist, otherwise an error could be returned. If the file can't be created or
// read, an error is returned as well.
func NewBBoltDAG(path string) (DAG, error) {
	if db, err := bbolt.Open(path, boltDBFileMode, bbolt.DefaultOptions); err != nil {
		return nil, fmt.Errorf("unable to create bbolt DAG: %w", err)
	} else {
		return &bboltDAG{db: db}, nil
	}
}

func (dag bboltDAG) MissingDocuments() []model.Hash {
	result := make([]model.Hash, 0)
	if err := dag.db.View(func(tx *bbolt.Tx) error {
		if bucket := tx.Bucket([]byte(missingDocumentsBucket)); bucket != nil {
			cursor := bucket.Cursor()
			for ref, _ := cursor.First(); ref != nil; ref, _ = cursor.Next() {
				result = append(result, model.SliceToHash(ref))
			}
		}
		return nil
	}); err != nil {
		log.Log().Errorf("Unable to fetch missing documents: %v", err)
	}
	return result
}

func (dag *bboltDAG) Add(documents ...Document) error {
	for _, document := range documents {
		if document != nil {
			if err := dag.add(document); err != nil {
				return err
			}
		}
	}
	return nil
}

func (dag bboltDAG) Root() (hash model.Hash, err error) {
	err = dag.db.View(func(tx *bbolt.Tx) error {
		if documents := tx.Bucket([]byte(documentsBucket)); documents != nil {
			if roots := getRoots(documents); len(roots) >= 1 {
				hash = roots[0]
			}
		}
		return nil
	})
	return
}

func (dag *bboltDAG) add(document Document) error {
	ref := document.Ref().Slice()
	return dag.db.Update(func(tx *bbolt.Tx) error {
		documents, nexts, missingDocuments, err := getBuckets(tx)
		if err != nil {
			return err
		}
		if exists(documents, document.Ref()) {
			log.Log().Tracef("Document %s already exists, not adding it again.", document.Ref())
			return nil
		}
		if len(document.Previous()) == 0 {
			if getRoots(documents) != nil {
				return errRootAlreadyExists
			}
			if err := addRoot(documents, document.Ref()); err != nil {
				return fmt.Errorf("unable to register root %s: %w", document.Ref(), err)
			}
		}
		if err := documents.Put(ref, document.Data()); err != nil {
			return err
		}
		// Store forward references ([C -> prev A, B] is stored as [A -> C, B -> C])
		for _, prev := range document.Previous() {
			if err := dag.registerNextRef(nexts, prev, document.Ref()); err != nil {
				return fmt.Errorf("unable to store forward reference %s->%s: %w", prev, document.Ref(), err)
			}
			if !exists(documents, prev) {
				log.Log().Debugf("Document %s is referring to missing prev %s, marking it as missing", document.Ref(), prev)
				if err = missingDocuments.Put(prev.Slice(), []byte{1}); err != nil {
					return fmt.Errorf("unable to register missing document %s: %w", prev, err)
				}
			}
		}
		// Remove marker if this document was previously missing
		return missingDocuments.Delete(ref)
	})
}

func getBuckets(tx *bbolt.Tx) (documents *bbolt.Bucket, nexts *bbolt.Bucket, missingDocuments *bbolt.Bucket, err error) {
	if documents, err = tx.CreateBucketIfNotExists([]byte(documentsBucket)); err != nil {
		return
	}
	if nexts, err = tx.CreateBucketIfNotExists([]byte(nextsBucket)); err != nil {
		return
	}
	if missingDocuments, err = tx.CreateBucketIfNotExists([]byte(missingDocumentsBucket)); err != nil {
		return
	}
	return
}

func getRoots(documentsBucket *bbolt.Bucket) []model.Hash {
	 return parseHashList(documentsBucket.Get([]byte(rootsDocumentKey)))
}

func addRoot(documentsBucket *bbolt.Bucket, ref model.Hash) error {
	roots := appendHashList(documentsBucket.Get([]byte(rootsDocumentKey)), ref)
	return documentsBucket.Put([]byte(rootsDocumentKey), roots)
}

// registerNextRef registers a forward reference a.k.a. "next", in contrary to "prev(s)" which is the inverse of the relation.
// It takes the nexts bucket, the prev and the next. Given document A and B where B prevs A, prev = A, next = B.
func (dag *bboltDAG) registerNextRef(nextsBucket *bbolt.Bucket, prev model.Hash, next model.Hash) error {
	prevSlice := prev.Slice()
	if value := nextsBucket.Get(prevSlice); value == nil {
		// No entry yet for this prev
		return nextsBucket.Put(prevSlice, next.Slice())
	} else {
		// Existing entry for this prev so add this one to it
		return nextsBucket.Put(prevSlice, appendHashList(value, next))
	}
}

func (dag bboltDAG) Walk(walker Walker, visitor Visitor, startAt model.Hash) error {
	return dag.db.View(func(tx *bbolt.Tx) error {
		documents := tx.Bucket([]byte(documentsBucket))
		nexts := tx.Bucket([]byte(nextsBucket))
		if documents == nil || nexts == nil {
			// DAG is empty
			return nil
		}
		return walker.walk(visitor, startAt, func(hash model.Hash) (Document, error) {
			documentBytes := documents.Get(hash.Slice())
			if documentBytes == nil {
				return nil, nil
			}
			if document, err := ParseDocument(documentBytes); err != nil {
				return nil, fmt.Errorf("unable to parse document %s: %w", hash, err)
			} else {
				return document, nil
			}
		}, func(hash model.Hash) ([]model.Hash, error) {
			return parseHashList(nexts.Get(hash.Slice())), nil
		}, documents.Stats().KeyN) // TODO Optimization: should we cache this number of keys?
	})
}

// exists checks whether the document with the given ref exists.
func exists(documents *bbolt.Bucket, ref model.Hash) bool {
	return documents.Get(ref.Slice()) != nil
}

// parseHashList splits a list of concatenated hashes into separate hashes.
func parseHashList(input []byte) []model.Hash {
	if len(input) == 0 {
		return nil
	}
	num := (len(input) - (len(input) % model.HashSize)) / model.HashSize
	result := make([]model.Hash, num)
	for i := 0; i < num; i++ {
		result[i] = model.SliceToHash(input[i*model.HashSize : i*model.HashSize+model.HashSize])
	}
	return result
}

func appendHashList(list []byte, hash model.Hash) []byte {
	newList := make([]byte, 0, len(list)+model.HashSize)
	newList = append(newList, list...)
	newList = append(newList, hash.Slice()...)
	return newList
}
