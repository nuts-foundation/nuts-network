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

package store

import (
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"io"
)

// DocumentStore is used to store and retrieve chained documents.
type DocumentStore interface {
	// Get retrieves a document from the store given its hash. If it doesn't exist, nil is returned (no error).
	Get(hash model.Hash) (*model.DocumentDescriptor, error)
	// GetByConsistencyHash retrieves a document from the store given its consistency hash. If it doesn't exist, nil is returned (no error).
	GetByConsistencyHash(hash model.Hash) (*model.DocumentDescriptor, error)
	// Add adds a document, returning the last consistency hash.
	Add(document model.Document) (model.Hash, error)
	// GetAll retrieves all documents.
	GetAll() ([]model.DocumentDescriptor, error)
	// WriteContents writes contents for the specified document, identified by the given hash.
	// If the document does not exist, an error is returned. If the document already has contents, it is overwritten.
	WriteContents(hash model.Hash, contents io.Reader) error
	// ReadContents reads the contents for the specified document, identified by the given hash. If the document doesn't exist, an error is returned.
	// If the document has no contents, nil is returned.
	ReadContents(hash model.Hash) (io.ReadCloser, error)
	// LastConsistencyHash retrieves the last consistency hash from the store (the consistency hash of the last document in the chain). If there are no documents in the store, an empty hash is returned.
	LastConsistencyHash() model.Hash
	// ContentsSize retrieves size of all stored contents in bytes.
	ContentsSize() int
	// Size retrieves the number of documents in the store.
	Size() int
}
