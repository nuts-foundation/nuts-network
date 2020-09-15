/*
 * Copyright (C) 2020. Nuts community
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 *
 */

package documentlog

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/nuts-foundation/nuts-network/pkg/documentlog/store"
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
		protocol := proto.NewMockProtocol(mockCtrl)
		documentStore := store.NewMockDocumentStore(mockCtrl)
		documentStore.EXPECT().Get(gomock.Any())
		documentStore.EXPECT().Add(gomock.Any())
		log := NewDocumentLog(protocol).(*documentLog)
		log.Configure(documentStore)
		expected := model.Document{
			Type:      "test",
			Timestamp: time.Now(),
		}
		expected.Hash = model.CalculateDocumentHash(expected.Type, expected.Timestamp, []byte{1, 2, 3})
		err := log.addDocument(expected)
		if !assert.NoError(t, err) {
			return
		}
	})
	t.Run("ok - multiple, random order", func(t *testing.T) {
		var protocol = proto.NewMockProtocol(mockCtrl)
		documentStore := store.NewMockDocumentStore(mockCtrl)
		log := NewDocumentLog(protocol).(*documentLog)
		log.Configure(documentStore)
		var count = 20
		documentStore.EXPECT().Get(gomock.Any()).Times(count)
		documentStore.EXPECT().Add(gomock.Any()).Times(count)
		for i := 0; i < count; i++ {
			doc := model.Document{
				Type:      fmt.Sprintf("%d", count),
				Timestamp: time.Now().AddDate(0, 0, rand.Intn(1000)),
			}
			doc.Hash = model.CalculateDocumentHash(doc.Type, doc.Timestamp, []byte{1, 2, 3})
			log.addDocument(doc)
		}
	})
	t.Run("ok - already exists", func(t *testing.T) {
		expected := model.DocumentDescriptor{
			Document: model.Document{
				Type:      "test",
				Timestamp: time.Now(),
			},
		}
		protocol := proto.NewMockProtocol(mockCtrl)
		documentStore := store.NewMockDocumentStore(mockCtrl)
		documentStore.EXPECT().Get(gomock.Any()).Return(&expected, nil)
		log := NewDocumentLog(protocol).(*documentLog)
		log.Configure(documentStore)
		err := log.addDocument(model.Document{Timestamp: expected.Document.Timestamp})
		assert.NoError(t, err)
	})
	t.Run("error - already exists, different timestamp", func(t *testing.T) {
		expected := model.DocumentDescriptor{
			Document: model.Document{
				Type:      "test",
				Timestamp: time.Now(),
			},
		}
		protocol := proto.NewMockProtocol(mockCtrl)
		documentStore := store.NewMockDocumentStore(mockCtrl)
		documentStore.EXPECT().Get(gomock.Any()).Return(&expected, nil)
		log := NewDocumentLog(protocol).(*documentLog)
		log.Configure(documentStore)
		err := log.addDocument(model.Document{})
		assert.Error(t, err)
	})
}

func Test_DocumentLog_AddDocumentWithContents(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	t.Run("ok - new document", func(t *testing.T) {
		protocol := proto.NewMockProtocol(mockCtrl)
		documentStore := store.NewMockDocumentStore(mockCtrl)
		documentStore.EXPECT().Get(gomock.Any()).Return(nil, nil)
		documentStore.EXPECT().Add(gomock.Any())
		documentStore.EXPECT().WriteContents(gomock.Any(), gomock.Any())
		storedDocument := model.DocumentDescriptor{
			Document: model.Document{
				Timestamp: time.Now(),
			},
		}
		documentStore.EXPECT().Get(gomock.Any()).Return(&storedDocument, nil)
		log := NewDocumentLog(protocol).(*documentLog)
		log.Configure(documentStore)
		expected := []byte{1, 2, 3}
		document, err := log.AddDocumentWithContents(storedDocument.Timestamp, "test", bytes.NewReader(expected))
		assert.NotNil(t, document)
		assert.NoError(t, err)
	})
}

func Test_DocumentLog_GetDocument(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	t.Run("ok - found", func(t *testing.T) {
		protocol := proto.NewMockProtocol(mockCtrl)
		documentStore := store.NewMockDocumentStore(mockCtrl)
		documentStore.EXPECT().Get(gomock.Any()).Return(&model.DocumentDescriptor{}, nil)
		log := NewDocumentLog(protocol).(*documentLog)
		log.Configure(documentStore)
		doc, err := log.GetDocument(model.EmptyHash())
		assert.NoError(t, err)
		assert.NotNil(t, doc)
	})
	t.Run("ok - not found", func(t *testing.T) {
		protocol := proto.NewMockProtocol(mockCtrl)
		documentStore := store.NewMockDocumentStore(mockCtrl)
		documentStore.EXPECT().Get(gomock.Any()).Return(nil, nil)
		log := NewDocumentLog(protocol).(*documentLog)
		log.Configure(documentStore)
		doc, err := log.GetDocument(model.EmptyHash())
		assert.Nil(t, doc)
		assert.NoError(t, err)
	})
	t.Run("error", func(t *testing.T) {
		protocol := proto.NewMockProtocol(mockCtrl)
		documentStore := store.NewMockDocumentStore(mockCtrl)
		documentStore.EXPECT().Get(gomock.Any()).Return(nil, errors.New("failed"))
		log := NewDocumentLog(protocol).(*documentLog)
		log.Configure(documentStore)
		doc, err := log.GetDocument(model.EmptyHash())
		assert.EqualError(t, err, "failed")
		assert.Nil(t, doc)
	})
}

func Test_DocumentLog_AddDocumentContents(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	t.Run("ok - already have contents", func(t *testing.T) {
		protocol := proto.NewMockProtocol(mockCtrl)
		documentStore := store.NewMockDocumentStore(mockCtrl)
		documentStore.EXPECT().Get(gomock.Any()).Return(&model.DocumentDescriptor{HasContents: true}, nil)
		log := NewDocumentLog(protocol).(*documentLog)
		log.Configure(documentStore)
		doc, err := log.AddDocumentContents(model.EmptyHash(), bytes.NewReader([]byte{1, 2, 3}))
		assert.NoError(t, err)
		assert.NotNil(t, doc)
	})
	t.Run("error - document not found", func(t *testing.T) {
		protocol := proto.NewMockProtocol(mockCtrl)
		documentStore := store.NewMockDocumentStore(mockCtrl)
		documentStore.EXPECT().Get(gomock.Any()).Return(nil, nil)
		log := NewDocumentLog(protocol).(*documentLog)
		log.Configure(documentStore)
		_, err := log.AddDocumentContents(model.EmptyHash(), bytes.NewReader([]byte{1, 2, 3}))
		assert.Error(t, err)
	})
}

func Test_DocumentLog_Subscribe(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	t.Run("ok", func(t *testing.T) {
		log := NewDocumentLog(proto.NewMockProtocol(mockCtrl)).(*documentLog)
		queue := log.Subscribe("some-type")
		assert.NotNil(t, queue)
		assert.Len(t, log.subscriptions, 1)
	})
}

func Test_DocumentLog_HasContents(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	t.Run("ok - document doesn't exist", func(t *testing.T) {
		protocol := proto.NewMockProtocol(mockCtrl)
		documentStore := store.NewMockDocumentStore(mockCtrl)
		documentStore.EXPECT().Get(gomock.Any()).Return(nil, nil)
		log := NewDocumentLog(protocol).(*documentLog)
		log.Configure(documentStore)
		result, err := log.HasContentsForDocument(model.EmptyHash())
		assert.NoError(t, err)
		assert.False(t, result)
	})
	t.Run("ok - document exists, no content", func(t *testing.T) {
		protocol := proto.NewMockProtocol(mockCtrl)
		documentStore := store.NewMockDocumentStore(mockCtrl)
		documentStore.EXPECT().Get(gomock.Any()).Return(&model.DocumentDescriptor{HasContents: false}, nil)
		log := NewDocumentLog(protocol).(*documentLog)
		log.Configure(documentStore)
		result, err := log.HasContentsForDocument(model.EmptyHash())
		assert.NoError(t, err)
		assert.False(t, result)
	})
	t.Run("ok - document exists, has content", func(t *testing.T) {
		protocol := proto.NewMockProtocol(mockCtrl)
		documentStore := store.NewMockDocumentStore(mockCtrl)
		documentStore.EXPECT().Get(gomock.Any()).Return(&model.DocumentDescriptor{HasContents: true}, nil)
		log := NewDocumentLog(protocol).(*documentLog)
		log.Configure(documentStore)
		result, err := log.HasContentsForDocument(model.EmptyHash())
		assert.NoError(t, err)
		assert.True(t, result)
	})
	t.Run("error", func(t *testing.T) {
		protocol := proto.NewMockProtocol(mockCtrl)
		documentStore := store.NewMockDocumentStore(mockCtrl)
		documentStore.EXPECT().Get(gomock.Any()).Return(nil, errors.New("failed"))
		log := NewDocumentLog(protocol).(*documentLog)
		log.Configure(documentStore)
		result, err := log.HasContentsForDocument(model.EmptyHash())
		assert.Error(t, err)
		assert.False(t, result)
	})
}

func Test_DocumentLog_Diagnostics(t *testing.T) {
	t.Run("ok - test race conditions (run with -race)", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		protocol := proto.NewMockProtocol(mockCtrl)
		documentStore := store.NewMockDocumentStore(mockCtrl)
		documentStore.EXPECT().Size()
		documentStore.EXPECT().ContentsSize()
		documentStore.EXPECT().Get(gomock.Any())
		documentStore.EXPECT().Add(gomock.Any())
		log := NewDocumentLog(protocol).(*documentLog)
		log.Configure(documentStore)
		go func() {
			log.addDocument(model.Document{})
		}()
		log.Statistics() // this should trigger a race condition if we had no locks
		time.Sleep(time.Second / 10)
	})
}
