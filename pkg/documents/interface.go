package documents

type DocumentLog interface {
	// Starts the document log
	Start()
	Stop()
}