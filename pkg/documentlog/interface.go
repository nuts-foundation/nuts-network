package documentlog

import (
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/proto"
)

type DocumentLog interface {
	proto.HashSource
	// Starts the document log
	Start()
	Stop()
	Subscribe(documentType string) DocumentQueue // TODO: Subscribe is a bad name when returning a blocking queue

	Diagnostics() []core.DiagnosticResult
}

type DocumentQueue interface {
	Get() *model.Document
}
