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
	"context"
	"encoding/json"
	"fmt"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// HttpClient holds the server address and other basic settings for the http client
type HttpClient struct {
	ServerAddress string
	Timeout       time.Duration
}

func (hb HttpClient) GetDocumentContents(hash model.Hash) (io.ReadCloser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()
	res, err := hb.client().GetDocumentContents(ctx, hash.String())
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if err := testResponseCode(http.StatusOK, res); err != nil {
		return nil, err
	}
	return res.Body, nil
}

func (hb HttpClient) ListDocuments() ([]model.Document, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()
	res, err := hb.client().ListDocuments(ctx)
	if err != nil {
		return nil, err
	}
	if err := testResponseCode(http.StatusOK, res); err != nil {
		return nil, err
	}
	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	documents := make([]Document, 0)
	if err := json.Unmarshal(responseData, &documents); err != nil {
		return nil, err
	}
	result := make([]model.Document, len(documents))
	for i, current := range documents {
		result[i] = current.toModel()
	}
	return result, nil
}

func (hb HttpClient) GetDocument(hash model.Hash) (*model.Document, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()
	res, err := hb.client().GetDocument(ctx, hash.String())
	if err != nil {
		return nil, err
	}
	return testAndParseDocumentResponse(res)
}

func (hb HttpClient) AddDocumentWithContents(timestamp time.Time, docType string, contents []byte) (model.Document, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()
	res, err := hb.client().AddDocument(ctx, AddDocumentJSONRequestBody{
		Contents:  contents,
		Timestamp: timestamp.UnixNano(),
		Type:      docType,
	})
	if err != nil {
		return model.Document{}, err
	}
	documentPtr, err := testAndParseDocumentResponse(res)
	if err != nil || documentPtr == nil {
		return model.Document{}, err
	}
	return *documentPtr, err
}

func (hb HttpClient) client() ClientInterface {
	url := hb.ServerAddress
	if !strings.Contains(url, "http") {
		url = fmt.Sprintf("http://%v", hb.ServerAddress)
	}

	response, err := NewClientWithResponses(url)
	if err != nil {
		panic(err)
	}
	return response
}

func testAndParseDocumentResponse(response *http.Response) (*model.Document, error) {
	if response.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if err := testResponseCode(http.StatusOK, response); err != nil {
		return nil, err
	}
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	result := Document{}
	if err := json.Unmarshal(responseData, &result); err != nil {
		return nil, err
	}
	doc := result.toModel()
	return &doc, nil
}

func testResponseCode(expectedStatusCode int, response *http.Response) error {
	if response.StatusCode != expectedStatusCode {
		responseData, _ := ioutil.ReadAll(response.Body)
		return fmt.Errorf("network returned HTTP %d (expected: %d), response: %s",
			response.StatusCode, expectedStatusCode, string(responseData))
	}
	return nil
}

func (d *Document) fromModel(document model.Document) {
	d.Timestamp = document.Timestamp.UnixNano()
	d.Hash = document.Hash.String()
	d.Type = document.Type
}

func (d Document) toModel() model.Document {
	hash, _ := model.ParseHash(d.Hash)
	return model.Document{
		Hash:      hash,
		Type:      d.Type,
		Timestamp: time.Unix(0, d.Timestamp),
	}
}
