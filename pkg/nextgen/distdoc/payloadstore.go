package distdoc

import (
	"github.com/nuts-foundation/nuts-network/pkg/model"
)

type PayloadStore interface {
	WritePayload(documentRef model.Hash, data []byte) error
	ReadPayload(documentRef model.Hash) ([]byte, error)
}
