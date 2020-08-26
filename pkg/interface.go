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

package pkg

import (
	"github.com/nuts-foundation/nuts-network/pkg/documentlog"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"io"
	"time"
)

// NetworkClient is the interface to be implemented by any remote or local client
type NetworkClient interface {
	documentlog.Publisher
	// GetDocumentContents retrieves the document contents for the given hash. If the document of contents are not known, an error is returned.
	GetDocumentContents(hash model.Hash) (io.ReadCloser, error)
	// GetDocument retrieves the document for the given hash. If the document is not known, an error is returned.
	GetDocument(hash model.Hash) (*model.DocumentDescriptor, error)
	// AddDocumentWithContents adds a document with the specified contents. If it already exists (hashes match) an error is returned.
	AddDocumentWithContents(timestamp time.Time, docType string, contents []byte) (*model.Document, error)
	// ListDocuments returns all documents known to this network instance.
	ListDocuments() ([]model.DocumentDescriptor, error)
}
