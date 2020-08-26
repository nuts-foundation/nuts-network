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
	"github.com/nuts-foundation/nuts-network/pkg/documentlog/store"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/proto"
	"github.com/nuts-foundation/nuts-network/pkg/stats"
)

// DocumentLog defines the API for the DocumentLog layer, used to store/retrieve chained documents.
type DocumentLog interface {
	Publisher
	proto.HashSource
	stats.StatsProvider
	// Configure configures this DocumentLog. Must be called before Start().
	Configure(store store.DocumentStore)
	// Starts the document log.
	Start()
	// Stops the document log.
	Stop()
}

type Publisher interface {
	// Subscribe creates a subscription for incoming documents with the specified type. It can be read from using the
	// returned DocumentQueue. There can be multiple subscribers for the same document type. The returned queue MUST
	// read from since it has an internal buffer which blocks the producer (the DocumentLog) when full.
	Subscribe(documentType string) DocumentQueue // TODO: Subscribe is a bad name when returning a blocking queue
}

// DocumentQueue is a blocking queue which allows callers to wait for documents to come in.
type DocumentQueue interface {
	// Get gets a document from the queue. It blocks until:
	// - There's a document to return
	// - The queue is closed by the producer
	// When the queue is closed this function an error and no document.
	Get() *model.Document
}

type documentQueue struct {
	c            chan *model.Document
	documentType string
}

func (q documentQueue) Get() *model.Document {
	return <-q.c
}
