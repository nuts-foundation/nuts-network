package documentlog

import (
	"fmt"
	"strings"
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

type ConsistencyHashListDiagnostic struct {
	Hashes []string
}

func (d ConsistencyHashListDiagnostic) Name() string {
	return "[DocumentLog] Ordered consistency hash list"
}

func (d ConsistencyHashListDiagnostic) String() string {
	var result = make([]string, len(d.Hashes))
	for i, hash := range d.Hashes {
		result[i] = fmt.Sprintf("%d->%s", i, hash)
	}
	return strings.Join(result, " ")
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
