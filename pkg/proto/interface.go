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
	"errors"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/p2p"
	"github.com/nuts-foundation/nuts-network/pkg/stats"
	"io"
	"time"
)

// Version holds the number of the version of this protocol implementation.
const Version = 1

// ErrMissingProtocolVersion is used when a message is received without protocol version.
var ErrMissingProtocolVersion = errors.New("missing protocol version")

// ErrUnsupportedProtocolVersion is used when a message is received with an unsupported protocol version.
var ErrUnsupportedProtocolVersion = errors.New("unsupported protocol version")

// Protocol defines the API for the protocol layer, which is a high-level interface to interact with the network. It responds
// from (peer) messages received through the P2P layer.
type Protocol interface {
	stats.StatsProvider
	// Configure configures the Protocol. Must be called before Start().
	Configure(p2pNetwork p2p.P2PNetwork, source HashSource)
	// Starts the Protocol (sending and receiving of messages).
	Start()
	// Stops the Protocol.
	Stop()
	// ReceivedConsistencyHashes returns a queue with consistency hashes we received from our peers. It must be drained, because when its buffer is full the producer is blocked.
	ReceivedConsistencyHashes() PeerHashQueue
	// AdvertConsistencyHash is used to tell our peers of our last consistency hash, so they can match their DocumentLog with ours (so replication can occur).
	AdvertConsistencyHash(hash model.Hash)
	// QueryHashList is used to query a peer for their document hash list.
	QueryHashList(peer model.PeerID) error
}

// PeerHashQueue is a queue which contains the hashes adverted by our peers. It's a FILO queue, since
// the hashes represent append-only data structures which means the last one is most recent.
type PeerHashQueue interface {
	// Get blocks until there's an PeerHash available and returns it.
	Get() *PeerHash
}

// PeerHash describes a hash we received from a peer.
type PeerHash struct {
	// Peer holds the ID of the peer we got the hash from.
	Peer model.PeerID
	// Hash holds the hash we received.
	Hash model.Hash
}

// HashSource is an SPI (Service Provider Interface) defined by the Protocol layer which is used by Protocol.
// It exists to break the circular dependency between the Protocol and DocumentLog layer.
type HashSource interface {
	// Documents retrieves all documents in the HashSource.
	Documents() ([]model.DocumentDescriptor, error)
	// AddDocument adds a document.
	AddDocument(model.Document) error
	// HasContentsForDocument determines whether we have the contents for the specified document. If we have the contents,
	// true is returned. If we don't have the contents or don't know the hash at all, false is returned.
	HasContentsForDocument(hash model.Hash) (bool, error)
	// GetDocument retrieves the document given the specified hash. If the document is not known, ErrUnknownDocument is returned.
	GetDocument(hash model.Hash) (*model.DocumentDescriptor, error)
	// GetDocumentContents retrieves the contents of the document given the specified hash. If the document is not known or we don't have its contents, ErrMissingDocumentContents is returned.
	GetDocumentContents(hash model.Hash) (io.ReadCloser, error)
	// FindByContentHash searches for documents which contents match the given SHA-1 hash.
	FindByContentHash(hash model.Hash) ([]model.DocumentDescriptor, error)
	// AddDocumentWithContents adds a document including contents.
	AddDocumentWithContents(timestamp time.Time, documentType string, contents io.Reader) (*model.Document, error)
	// AddDocumentContents adds a contents to an already known document.
	AddDocumentContents(hash model.Hash, contents io.Reader) (*model.Document, error)
}

type chanPeerHashQueue struct {
	c chan *PeerHash
}

func (q chanPeerHashQueue) Get() *PeerHash {
	return <-q.c
}
