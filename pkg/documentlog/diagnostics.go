package documentlog

import (
	"fmt"
)

type LastConsistencyHashDiagnostic struct {
	Hash string
}

func (d LastConsistencyHashDiagnostic) Name() string {
	return "[DocumentLog] Last consistency hash"
}

func (d LastConsistencyHashDiagnostic) String() string {
	return d.Hash
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
	return "[DocumentLog] Stored document size (bytes)"
}

func (d LogSizeDiagnostic) String() string {
	return fmt.Sprintf("%d", d.SizeInBytes)
}
