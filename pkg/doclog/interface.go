package doclog

import "github.com/nuts-foundation/nuts-network/pkg/model"

type DocumentLog interface {
	// Starts the document log
	Start()
	Stop()
	Add(document *model.Document)
}