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

package api

import (
	"encoding/json"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/test"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type handler struct {
	statusCode   int
	responseData []byte
}

func (h handler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	writer.WriteHeader(h.statusCode)
	writer.Write(h.responseData)
}

func TestDocument_fromModel(t *testing.T) {
	expected := model.Document{
		Hash:      test.StringToHash("37076f2cbe014a109d79b61ae9c7191f4cc57afc"),
		Type:      "test",
		Timestamp: time.Now(),
	}
	actual := Document{}
	actual.fromModel(expected)
	assert.Equal(t, actual.Type, expected.Type)
	assert.Equal(t, actual.Hash, expected.Hash.String())
	assert.Equal(t, actual.Timestamp, expected.Timestamp.UnixNano())
}

func TestDocument_toModel(t *testing.T) {
	expected := model.Document{
		Hash:      test.StringToHash("37076f2cbe014a109d79b61ae9c7191f4cc57afc"),
		Type:      "test",
		Timestamp: time.Unix(0, 1591255458635412720).UTC(),
	}
	actual := Document{
		Hash:      "37076f2cbe014a109d79b61ae9c7191f4cc57afc",
		Timestamp: 1591255458635412720,
		Type:      "test",
	}
	assert.Equal(t, expected, actual.toModel())
}

func TestHttpClient_ListDocuments(t *testing.T) {
	t.Run("200", func(t *testing.T) {
		doc1 := Document{
			Hash:      "31a3d460bb3c7d98845187c716a30db81c44b615",
			Timestamp: 1000,
			Type:      "doc",
		}
		doc2 := Document{
			Hash:      "bdeab373c65ef320514763bc94eaa827bee14915",
			Timestamp: 2000,
			Type:      "doc",
		}
		data, _ := json.Marshal([]Document{doc1, doc2})
		s := httptest.NewServer(handler{statusCode: http.StatusOK, responseData: data})
		httpClient := HttpClient{ServerAddress: s.URL, Timeout: time.Second}
		documents, err := httpClient.ListDocuments()
		if !assert.NoError(t, err) {
			return
		}
		if !assert.Len(t, documents, 2) {
			return
		}
		assert.Equal(t, doc1.Hash, documents[0].Hash.String())
		assert.Equal(t, doc1.Timestamp, documents[0].Timestamp.UnixNano())
		assert.Equal(t, doc1.Type, documents[0].Type)
		assert.Equal(t, doc2.Hash, documents[1].Hash.String())
		assert.Equal(t, doc2.Timestamp, documents[1].Timestamp.UnixNano())
		assert.Equal(t, doc2.Type, documents[1].Type)
	})
}

func TestHttpClient_Subscribe(t *testing.T) {
	t.Run("not supported", func(t *testing.T) {
		httpClient := HttpClient{ServerAddress: "http://foo", Timeout: time.Second}
		queue := httpClient.Subscribe("docType")
		assert.Nil(t, queue)
	})
}
