package pkg

import (
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"io"
	"time"
)

// NetworkClient is the interface to be implemented by any remote or local client
type NetworkClient interface {

	GetDocumentContents(hash model.Hash) (io.ReadCloser, error)
	GetDocument(hash model.Hash) (*model.Document, error)
	AddDocumentWithContents(timestamp time.Time, docType string, contents []byte) (model.Document, error)
	ListDocuments() ([]model.Document, error)
}
