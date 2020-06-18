package documentlog

import (
	"bytes"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/proto"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func Test_DocumentLog_AddDocument(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	t.Run("ok", func(t *testing.T) {
		var protocol = proto.NewMockProtocol(mockCtrl)
		log := NewDocumentLog(protocol)
		expected := model.Document{
			Type:      "test",
			Timestamp: time.Now(),
		}
		expected.Hash = model.CalculateDocumentHash(expected.Type, expected.Timestamp, []byte{1, 2, 3})
		log.AddDocument(expected)
		docs, _ := log.Documents()
		assert.Len(t, docs, 1)
		hasDoc, _ := log.HasDocument(expected.Hash)
		assert.True(t, hasDoc)
		hasContents, _ := log.HasContentsForDocument(expected.Hash)
		assert.False(t, hasContents)
	})
	t.Run("ok - multiple, random order", func(t *testing.T) {
		var protocol = proto.NewMockProtocol(mockCtrl)
		log := NewDocumentLog(protocol)
		var count = 20
		for i := 0; i < count; i++ {
			doc := model.Document{
				Type:      fmt.Sprintf("%d", count),
				Timestamp: time.Now().AddDate(0, 0, rand.Intn(1000)),
			}
			doc.Hash = model.CalculateDocumentHash(doc.Type, doc.Timestamp, []byte{1, 2, 3})
			log.AddDocument(doc)
		}
		docs, _ := log.Documents()
		assert.Len(t, docs, count)
	})
	t.Run("ok - already exists", func(t *testing.T) {
		var protocol = proto.NewMockProtocol(mockCtrl)
		log := NewDocumentLog(protocol)
		expected := model.Document{
			Type:      "test",
			Timestamp: time.Now(),
		}
		expected.Hash = model.CalculateDocumentHash(expected.Type, expected.Timestamp, []byte{1, 2, 3})
		docs, _ := log.Documents()
		assert.Len(t, docs, 0)
		log.AddDocument(expected)
		log.AddDocument(expected)
		docs, _ = log.Documents()
		assert.Len(t, docs, 1)
	})
	t.Run("error - already exists, different timestamp", func(t *testing.T) {
		var protocol = proto.NewMockProtocol(mockCtrl)
		log := NewDocumentLog(protocol)
		expected := model.Document{
			Type:      "test",
			Timestamp: time.Now(),
		}
		expected.Hash = model.CalculateDocumentHash(expected.Type, expected.Timestamp, []byte{1, 2, 3})
		documents, _ := log.Documents()
		assert.Len(t, documents, 0)
		log.AddDocument(expected)
		expected.Timestamp = time.Time{}
		log.AddDocument(expected)
		documents, _ = log.Documents()
		assert.Len(t, documents, 1)
	})
}

func Test_DocumentLog_AddDocumentWithContents(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	t.Run("ok", func(t *testing.T) {
		var protocol = proto.NewMockProtocol(mockCtrl)
		log := NewDocumentLog(protocol)
		timestamp := time.Now()
		expected := []byte{1, 2, 3}
		document, err := log.AddDocumentWithContents(timestamp, "test", bytes.NewReader(expected))
		assert.NoError(t, err)
		assert.Equal(t, timestamp, document.Timestamp)
		assert.Equal(t, "test", document.Type)
		documents, _ := log.Documents()
		assert.Len(t, documents, 1)
		// Assert that we can retrieve the contents
		actual, err := log.GetDocumentContents(document.Hash)
		if !assert.NoError(t, err) {
			return
		}
		buffer := new(bytes.Buffer)
		buffer.ReadFrom(actual)
		assert.Equal(t, expected, buffer.Bytes())
	})
	t.Run("error - contents already exist", func(t *testing.T) {
		var protocol = proto.NewMockProtocol(mockCtrl)
		log := NewDocumentLog(protocol)
		timestamp := time.Now()
		log.AddDocumentWithContents(timestamp, "test", bytes.NewReader([]byte{1, 2, 3}))
		_, err := log.AddDocumentWithContents(timestamp, "test", bytes.NewReader([]byte{1, 2, 3}))
		assert.Error(t, err)
	})
}

func Test_DocumentLog_AddDocumentContents(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	t.Run("error - document not found", func(t *testing.T) {
		var protocol = proto.NewMockProtocol(mockCtrl)
		log := NewDocumentLog(protocol)
		_, err := log.AddDocumentContents(model.EmptyHash(), bytes.NewReader([]byte{1, 2, 3}))
		assert.Error(t, err)
	})
}

func Test_DocumentLog_Diagnostics(t *testing.T) {
	t.Run("ok - test race conditions (run with -race)", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		var protocol = proto.NewMockProtocol(mockCtrl)
		log := NewDocumentLog(protocol)
		go func() {
			log.AddDocumentWithContents(time.Now(), "test", bytes.NewReader([]byte{1, 2, 3}))
		}()
		log.Statistics() // this should trigger a race condition if we had no locks
	})
}