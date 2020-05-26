package model

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"time"
)

const HashSize = 20

type DocumentHash struct {
	Hash
	Timestamp time.Time
}

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

// TODO: Unused?
func (h Hash) Equals(other Hash) bool {
	return bytes.Compare(h, other) == 0
}

func (h Hash) String() string {
	return hex.EncodeToString(h)
}

func MakeConsistencyHash(target Hash, h1 Hash, h2 Hash) {
	// TODO: This a naive, relatively slow to XOR 2 byte slices. There's a faster way: https://github.com/lukechampine/fastxor/blob/master/xor.go
	for i, _ := range h1 {
		target[i] = h1[i] ^ h2[i]
	}
}

type NodeID string

func (n NodeID) String() string {
	return string(n)
}

type PeerID string

func GetPeerID(addr string) PeerID {
	return PeerID(addr)
}

// TODO: Should this be an interface?
type Document struct {
	Contents  []byte
	Timestamp time.Time
}

func (d Document) Hash() Hash {
	h := sha1.Sum(d.Contents)
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
