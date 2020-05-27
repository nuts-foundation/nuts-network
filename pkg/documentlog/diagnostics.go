package documentlog

import (
	"fmt"
	"github.com/nuts-foundation/nuts-network/pkg/model"
)

type LastConsistencyHashDiagnostic struct {
	Hash model.Hash
}

func (d LastConsistencyHashDiagnostic) Name() string {
	return "[DocumentLog] Last consistency hash"
}

func (d LastConsistencyHashDiagnostic) String() string {
	return d.Hash.String()
}

type NumberOfDocumentsDiagnostic struct {
	NumberOfDocuments int
}

func (d NumberOfDocumentsDiagnostic) Name() string {
	return "[DocumentLog] Number of documents"
}

func (d NumberOfDocumentsDiagnostic) String() string {
	return fmt.Sprintf("%d", d.NumberOfDocuments)
}

type LogSizeDiagnostic struct {
	SizeInBytes int
}

func (d LogSizeDiagnostic) Name() string {
	return "[DocumentLog] Stored document size"
}

func (d LogSizeDiagnostic) String() string {
	return fmt.Sprintf("%d", d.SizeInBytes)
}
