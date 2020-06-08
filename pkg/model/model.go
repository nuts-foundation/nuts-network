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

package model

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

const HashSize = 20

type Hash []byte

func EmptyHash() Hash {
	return make([]byte, HashSize)
}

func (h Hash) Empty() bool {
	// TODO: Isn't this a bit slow?
	for _, b := range h {
		if b != 0 {
			return false
		}
	}
	return true
}

func (h Hash) Clone() Hash {
	clone := EmptyHash()
	copy(clone, h)
	return clone
}

func (h Hash) Equals(other Hash) bool {
	return bytes.Compare(h, other) == 0
}

func (h Hash) String() string {
	return hex.EncodeToString(h)
}

func ParseHash(input string) (Hash, error) {
	bytes, err := hex.DecodeString(input)
	if err != nil {
		return nil, err
	}
	if len(bytes) != HashSize {
		return nil, fmt.Errorf("incorrect hash length (%d)", len(bytes))
	}
	return bytes, nil
}

func MakeConsistencyHash(h1 Hash, h2 Hash) Hash {
	target := EmptyHash()
	// TODO: This a naive, relatively slow to XOR 2 byte slices. There's a faster way: https://github.com/lukechampine/fastxor/blob/master/xor.go
	for i, _ := range h1 {
		target[i] = h1[i] ^ h2[i]
	}
	return target
}

type NodeID string

func (n NodeID) Empty() bool {
	return n == ""
}

func (n NodeID) String() string {
	return string(n)
}

type PeerID string

func GetPeerID(addr string) PeerID {
	return PeerID(addr)
}

// TODO: Should this be an interface?
type Document struct {
	Hash      Hash
	Type      string
	Timestamp time.Time
}

func CalculateDocumentHash(docType string, timestamp time.Time, contents []byte) Hash {
	// TODO: Document this
	input := map[string]interface{}{
		"contents":  contents,
		"type":      docType,
		"timestamp": timestamp.UnixNano(),
	}
	// TODO: Canonicalize?
	data, _ := json.Marshal(input)
	h := sha1.Sum(data)
	return h[:]
}

type NodeInfo struct {
	ID      NodeID
	Address string
}

func (n NodeInfo) String() string {
	return fmt.Sprintf("%s(%s)", n.ID, n.Address)
}

func ParseNodeInfo(addr string) NodeInfo {
	return NodeInfo{Address: addr}
}
