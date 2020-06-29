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
	"github.com/labstack/echo/v4"
	logging "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/pkg"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// ApiWrapper is needed to connect the implementation to the echo ServiceWrapper
type ApiWrapper struct {
	Service pkg.NetworkClient
}

func (a ApiWrapper) ListDocuments(ctx echo.Context) error {
	documents, err := a.Service.ListDocuments()
	if err != nil {
		logging.Log().Errorf("Error while listing documents: %v", err)
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	results := make([]Document, len(documents))
	for i, document := range documents {
		results[i] = Document{}
		results[i].fromModel(document.Document)
	}
	return ctx.JSON(http.StatusOK, results)
}

func (a ApiWrapper) GetDocument(ctx echo.Context, hashAsString string) error {
	hash, err := model.ParseHash(hashAsString)
	if err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	document, err := a.Service.GetDocument(hash)
	if err != nil {
		logging.Log().Errorf("Error while retrieving document (hash=%s): %v", hash, err)
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	if document == nil {
		return ctx.String(http.StatusNotFound, "document not found")
	}
	return writeDocument(ctx, http.StatusOK, document.Document)
}

func (a ApiWrapper) GetDocumentContents(ctx echo.Context, hashAsString string) error {
	hash, err := model.ParseHash(hashAsString)
	if err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	reader, err := a.Service.GetDocumentContents(hash)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	if reader == nil {
		return ctx.String(http.StatusNotFound, "document or contents not found")
	}
	defer reader.Close()
	ctx.Response().Header().Set(echo.HeaderContentType, "application/octet-stream")
	ctx.Response().WriteHeader(http.StatusOK)
	_, err = io.Copy(ctx.Response().Writer, reader)
	return err
}

func (a ApiWrapper) AddDocument(ctx echo.Context) error {
	inputDocument, err := parseDocument(ctx)
	if err != nil {
		logging.Log().Warnf("Unable to parse document: %v", err)
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	outputDocument, err := a.Service.AddDocumentWithContents(time.Unix(0, inputDocument.Timestamp), inputDocument.Type, inputDocument.Contents)
	if err != nil {
		logging.Log().Warnf("Unable to add document: %v", err)
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	return writeDocument(ctx, http.StatusCreated, *outputDocument)
}

func writeDocument(ctx echo.Context, statusCode int, document model.Document) error {
	return ctx.JSON(statusCode, Document{
		Hash:      document.Hash.String(),
		Timestamp: document.Timestamp.UnixNano(),
		Type:      document.Type,
	})
}

func parseDocument(ctx echo.Context) (DocumentWithContents, error) {
	bytes, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		return DocumentWithContents{}, err
	}
	document := DocumentWithContents{}
	err = json.Unmarshal(bytes, &document)
	if err != nil {
		return DocumentWithContents{}, err
	}
	return document, err
}
