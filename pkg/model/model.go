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

// model contains data structures canonical to all layers in nuts-network
package model

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// HashSize holds the size of hashes used on the network.
const HashSize = 20

// Hash is a type that holds a hash used on the network.
type Hash [HashSize]byte

// EmptyHash returns a Hash that is empty (initialized with zeros).
func EmptyHash() Hash {
	return [HashSize]byte{}
}

// Empty tests whether the Hash is empty (all zeros).
func (h Hash) Empty() bool {
	// TODO: Isn't this a bit slow?
	for _, b := range h {
		if b != 0 {
			return false
		}
	}
	return true
}

// Clone returns a copy of the Hash.
func (h Hash) Clone() Hash {
	clone := EmptyHash()
	copy(clone[:], h[:])
	return clone
}

// Slice returns the Hash as a slice. It does not copy the array.
func (h Hash) Slice() []byte {
	return h[:]
}

// Equals determines whether the given Hash is exactly the same (bytes match).
func (h Hash) Equals(other Hash) bool {
	return h.Compare(other) == 0
}

// Compare compares this Hash to another Hash using bytes.Compare.
func (h Hash) Compare(other Hash) int {
	return bytes.Compare(h[:], other[:])
}

// String returns the Hash in serialized (and human-readable) form.
func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

// SliceToHash converts a byte slice to a Hash, returning a copy.
func SliceToHash(slice []byte) Hash {
	result := EmptyHash()
	copy(result[:], slice)
	return result
}

// ParseHash parses the given input string as Hash. If the input is invalid and can't be parsed as Hash, an error is returned.
func ParseHash(input string) (Hash, error) {
	if input == "" {
		return EmptyHash(), nil
	}
	data, err := hex.DecodeString(input)
	if err != nil {
		return EmptyHash(), err
	}
	if len(data) != HashSize {
		return EmptyHash(), fmt.Errorf("incorrect hash length (%d)", len(data))
	}
	result := EmptyHash()
	copy(result[0:], data)
	return result, nil
}

// MakeConsistencyHash calculates a consistency hash for the given 2 input Hashes and returns it.
func MakeConsistencyHash(h1 Hash, h2 Hash) Hash {
	target := EmptyHash()
	// TODO: This a naive, relatively slow to XOR 2 byte slices. There's a faster way: https://github.com/lukechampine/fastxor/blob/master/xor.go
	for i := range h1 {
		target[i] = h1[i] ^ h2[i]
	}
	return target
}

// NodeID is a self-proclaimed unique identifier for nodes on the network.
type NodeID string

// Empty checks whether the NodeID is empty (empty string).
func (n NodeID) Empty() bool {
	return n == ""
}

// String returns the NodeID as string.
func (n NodeID) String() string {
	return string(n)
}

// PeerID identifies a peer (connected node) on the network
type PeerID string

// GetPeerID constructs a PeerID given its remote address.
func GetPeerID(addr string) PeerID {
	return PeerID(addr)
}

// DocumentDescriptor is just a Document with often-required properties added (HasContents indicator) to optimize
// access to the document storage.
type DocumentDescriptor struct {
	Document
	// HasContents indicates whether we have the contents of the document.
	HasContents bool
	// ConsistencyHash contains the actual consistency hash for the document.
	ConsistencyHash Hash
}

// Document describes a document on the network.
type Document struct {
	// Hash contains the hash of the document (which is calculated using its type, timestamp and contents).
	Hash Hash
	// Type contains a key describing what the document holds (e.g. 'nuts.node-info'). It is free format.
	Type string
	// Timestamp holds the moment the document was created. When serialized (for storage or transport on the network) it is converted to UTC and represented in Unix nanoseconds.
	Timestamp time.Time
}

// Clone makes a deep copy of the document and returns it.
func (d Document) Clone() Document {
	cp := d
	cp.Hash = cp.Hash.Clone()
	return cp
}

// CalculateDocumentHash calculates the hash for a Document in a canonicalized manner.
func CalculateDocumentHash(docType string, timestamp time.Time, contents []byte) Hash {
	data, _ := json.Marshal(struct {
		DocType   string `json:"type"`
		Timestamp int64  `json:"timestamp"`
		Contents  []byte `json:"contents"`
	}{docType, MarshalDocumentTime(timestamp), contents})
	return sha1.Sum(data)
}

// NodeInfo describes a known remote node on the network.
type NodeInfo struct {
	// ID holds the NodeID by which we know the node.
	ID NodeID
	// Address holds the remote address of the node.
	Address string
}

// String returns the NodeInfo in human-readable format.
func (n NodeInfo) String() string {
	return fmt.Sprintf("%s(%s)", n.ID, n.Address)
}
