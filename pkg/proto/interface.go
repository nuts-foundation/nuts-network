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

package proto

import (
	"context"
	"errors"
	"github.com/nuts-foundation/nuts-network/pkg/concurrency"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/p2p"
	"github.com/nuts-foundation/nuts-network/pkg/stats"
	"io"
	"time"
)

const Version = 1

var ErrMissingProtocolVersion = errors.New("missing protocol version")

var ErrUnsupportedProtocolVersion = errors.New("unsupported protocol version")

type Protocol interface {
	stats.StatsProvider
	// TODO: This feels like poor man's dependency injection... We need a design pattern here.
	Start(p2pNetwork p2p.P2PNetwork, source HashSource)
	Stop()

	ReceivedConsistencyHashes() PeerHashQueue
	ReceivedDocumentHashes() PeerHashQueue

	AdvertConsistencyHash(hash model.Hash)
	QueryHashList(peer model.PeerID) error
}

// PeerHashQueue is a queue which contains the hashes adverted by our peers. It's a FILO queue, since
// the hashes represent append-only data structures which means the last one is most recent.
type PeerHashQueue interface {
	// Get blocks until there's an PeerHash available and returns it.
	Get(cxt context.Context) (PeerHash, error)
}

type PeerHash struct {
	Peer model.PeerID
	Hash model.Hash
}

type HashSource interface {
	Documents() []model.Document
	// HasDocument tests whether the document is present for the given hash
	HasDocument(hash model.Hash) bool
	HasContentsForDocument(hash model.Hash) bool
	GetDocument(hash model.Hash) *model.Document
	GetDocumentContents(hash model.Hash) (io.ReadCloser, error)
	AddDocument(document model.Document)
	AddDocumentWithContents(timestamp time.Time, documentType string, contents io.Reader) (model.Document, error)
	AddDocumentContents(hash model.Hash, contents io.Reader) (model.Document, error)
}

type AdvertedHashQueue struct {
	internal concurrency.Queue
}

func (q AdvertedHashQueue) Get(cxt context.Context) (PeerHash, error) {
	item, err := q.internal.Get(cxt)
	if err != nil {
		return PeerHash{}, err
	} else {
		return item.(PeerHash), nil
	}
}
