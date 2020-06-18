package store

import (
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"io"
)

type DocumentStore interface {
	Get(hash model.Hash) (*StoredDocument, error)
	GetByConsistencyHash(hash model.Hash) (*StoredDocument, error)
	// Add adds a document, returning the last consistency hash
	Add(document model.Document) (model.Hash, error)
	GetAll() ([]StoredDocument, error)
	WriteContents(hash model.Hash, contents io.Reader) error
	ReadContents(hash model.Hash) (io.ReadCloser, error)
	LastConsistencyHash() model.Hash
	ContentsSize() int
	Size() int
}

type StoredDocument struct {
	HasContents bool
	model.Document
}
