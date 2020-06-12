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
	"context"
	"github.com/nuts-foundation/nuts-network/pkg/concurrency"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/proto"
	"github.com/nuts-foundation/nuts-network/pkg/stats"
)

type DocumentLog interface {
	proto.HashSource
	stats.StatsProvider
	// Starts the document log
	Start()
	Stop()
	Subscribe(documentType string) DocumentQueue // TODO: Subscribe is a bad name when returning a blocking queue
}

// DocumentQueue is a blocking queue which allows callers to wait for documents to come in.
type DocumentQueue interface {
	// Get gets a document from the queue. It blocks until:
	// - There's a document to return
	// - The context is cancelled or expires
	// - The queue is closed by the producer
	// When the queue is closed this function an error and no document.
	Get(context context.Context) (model.Document, error)
}

type documentQueue struct {
	internal     concurrency.Queue
	documentType string
}

func (q documentQueue) Get(cxt context.Context) (model.Document, error) {
	item, err := q.internal.Get(cxt)
	if err != nil {
		return model.Document{}, err
	} else {
		return item.(model.Document), nil
	}
}
