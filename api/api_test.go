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
 */

package api

import (
	"bytes"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/nuts-foundation/nuts-network/pkg"
	"github.com/nuts-foundation/nuts-network/pkg/documentlog"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/test"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var document = model.Document{
	Hash:      test.StringToHash("37076f2cbe014a109d79b61ae9c7191f4cc57afc"),
	Type:      "mock",
	Timestamp: time.Unix(0, 1591255458635412720),
}
var documentContents = []byte("foobar")

const documentAsJSON = "{\"hash\":\"37076f2cbe014a109d79b61ae9c7191f4cc57afc\",\"timestamp\":1591255458635412720,\"type\":\"mock\"}"

func TestApiWrapper_AddDocument(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	path := "/document/"
	t.Run("ok", func(t *testing.T) {
		var networkClient = pkg.NewMockNetworkClient(mockCtrl)
		e, wrapper := initMockEcho(networkClient)
		networkClient.EXPECT().AddDocumentWithContents(document.Timestamp, document.Type, documentContents).Return(document, nil)

		input := DocumentWithContents{
			Contents:  documentContents,
			Timestamp: document.Timestamp.UnixNano(),
			Type:      document.Type,
		}
		req := httptest.NewRequest(echo.POST, "/", &test.JSONReader{Input: input})
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath(path)

		err := wrapper.AddDocument(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.JSONEq(t, documentAsJSON, rec.Body.String())
	})
	t.Run("not found", func(t *testing.T) {
		var networkClient = pkg.NewMockNetworkClient(mockCtrl)
		e, wrapper := initMockEcho(networkClient)
		networkClient.EXPECT().GetDocumentContents(gomock.Any()).Return(nil, nil)

		req := httptest.NewRequest(echo.GET, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath(path)
		c.SetParamNames("hash")
		c.SetParamValues(document.Hash.String())

		err := wrapper.GetDocumentContents(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "document or contents not found", rec.Body.String())
	})
}

func TestApiWrapper_GetDocument(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	path := "/document/:hash"
	t.Run("ok", func(t *testing.T) {
		var networkClient = pkg.NewMockNetworkClient(mockCtrl)
		e, wrapper := initMockEcho(networkClient)
		networkClient.EXPECT().GetDocument(test.EqHash(document.Hash)).Return(&document, nil)

		req := httptest.NewRequest(echo.GET, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath(path)
		c.SetParamNames("hash")
		c.SetParamValues(document.Hash.String())

		err := wrapper.GetDocument(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, documentAsJSON, rec.Body.String())
	})
	t.Run("not found", func(t *testing.T) {
		var networkClient = pkg.NewMockNetworkClient(mockCtrl)
		e, wrapper := initMockEcho(networkClient)
		networkClient.EXPECT().GetDocument(gomock.Any()).Return(nil, nil)

		req := httptest.NewRequest(echo.GET, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath(path)
		c.SetParamNames("hash")
		c.SetParamValues(document.Hash.String())

		err := wrapper.GetDocument(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "document not found", rec.Body.String())
	})
}

func TestApiWrapper_GetDocumentContents(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	path := "/document/:hash/contents"
	t.Run("ok", func(t *testing.T) {
		var networkClient = pkg.NewMockNetworkClient(mockCtrl)
		e, wrapper := initMockEcho(networkClient)
		networkClient.EXPECT().GetDocumentContents(test.EqHash(document.Hash)).Return(documentlog.NoopCloser{Reader: bytes.NewReader(documentContents)}, nil)

		req := httptest.NewRequest(echo.GET, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath(path)
		c.SetParamNames("hash")
		c.SetParamValues(document.Hash.String())

		err := wrapper.GetDocumentContents(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "foobar", rec.Body.String())
	})
	t.Run("not found", func(t *testing.T) {
		var networkClient = pkg.NewMockNetworkClient(mockCtrl)
		e, wrapper := initMockEcho(networkClient)
		networkClient.EXPECT().GetDocumentContents(gomock.Any()).Return(nil, nil)

		req := httptest.NewRequest(echo.GET, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath(path)
		c.SetParamNames("hash")
		c.SetParamValues(document.Hash.String())

		err := wrapper.GetDocumentContents(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "document or contents not found", rec.Body.String())
	})
}

func TestApiWrapper_ListDocuments(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	t.Run("list documents", func(t *testing.T) {
		t.Run("200", func(t *testing.T) {
			var networkClient = pkg.NewMockNetworkClient(mockCtrl)
			e, wrapper := initMockEcho(networkClient)
			networkClient.EXPECT().ListDocuments().Return([]model.Document{document}, nil)

			req := httptest.NewRequest(echo.GET, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/document")

			err := wrapper.ListDocuments(c)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.JSONEq(t, "[{\"hash\":\"37076f2cbe014a109d79b61ae9c7191f4cc57afc\",\"timestamp\":1591255458635412720,\"type\":\"mock\"}]", rec.Body.String())
		})
	})
}

func initMockEcho(networkClient *pkg.MockNetworkClient) (*echo.Echo, *ServerInterfaceWrapper) {
	e := echo.New()
	stub := ApiWrapper{Service: networkClient}
	wrapper := &ServerInterfaceWrapper{
		Handler: stub,
	}
	return e, wrapper
}
