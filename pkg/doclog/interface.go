package doclog

import (
	"github.com/nuts-foundation/nuts-network/pkg/proto"
)

type DocumentLog interface {
	proto.HashSource
	// Starts the document log
	Start()
	Stop()
}