package store

import (
	"bytes"
	"errors"
	"fmt"
	log "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/pkg/concurrency"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"io"
	"sort"
	"sync/atomic"
)

func NewMemoryDocumentStore() DocumentStore {
	return &memoryDocumentStore{
		documentHashIndex:   make(map[string]*memoryEntry, 0),
		mutex:               concurrency.NewSaferRWMutex("doclog-docs"),
		lastConsistencyHash: &atomic.Value{},
		contentsSize:        new(int32),
	}
}

type memoryDocumentStore struct {
	// entries contains the list of entries in the document log
	entries []*memoryEntry
	// documentHashIndex contains the entries indexed by document hash
	documentHashIndex map[string]*memoryEntry
	// consistencyHashIndex contains the entries indexed by consistency hash
	consistencyHashIndex map[string]*memoryEntry
	mutex                concurrency.SaferRWMutex
	lastConsistencyHash  *atomic.Value
	contentsSize         *int32
}

func (m *memoryDocumentStore) GetByConsistencyHash(hash model.Hash) (*StoredDocument, error) {
	var result *StoredDocument
	m.mutex.ReadLock(func() {
		if entry := m.consistencyHashIndex[hash.String()]; entry != nil {
			result = &StoredDocument{
				HasContents: entry.contents != nil,
				Document:    entry.Document.Clone(),
			}
		}
	})
	return result, nil
}

func (m *memoryDocumentStore) GetAll() ([]StoredDocument, error) {
	result := make([]StoredDocument, 0)
	m.mutex.ReadLock(func() {
		for _, entry := range m.entries {
			result = append(result, StoredDocument{
				HasContents: entry.contents != nil,
				Document:    entry.Document.Clone(),
			})
		}
	})
	return result, nil
}

type memoryEntry struct {
	model.Document
	contents *[]byte
}

func (m *memoryDocumentStore) WriteContents(hash model.Hash, contents io.Reader) error {
	var err error
	m.mutex.WriteLock(func() {
		entry := m.documentHashIndex[hash.String()]
		if entry == nil {
			err = errors.New("unknown document")
			return
		}
		if entry.contents == nil {
			buffer := new(bytes.Buffer)
			_, err = buffer.ReadFrom(contents)
			if err != nil {
				return
			}
			bytes := buffer.Bytes()
			actualHash := model.CalculateDocumentHash(entry.Type, entry.Timestamp, bytes)
			if !entry.Hash.Equals(hash) {
				log.Log().Warnf("Document rejected, actual hash differs from expected (expected=%s,actual=%s)", entry.Hash, actualHash)
				return
			}
			entry.contents = &bytes
			atomic.AddInt32(m.contentsSize, int32(len(bytes)))
		}
	})
	return err
}

func (m memoryDocumentStore) ReadContents(hash model.Hash) (io.ReadCloser, error) {
	var entry *memoryEntry
	m.mutex.ReadLock(func() {
		entry = m.documentHashIndex[hash.String()]
	})
	if entry == nil || entry.contents == nil {
		return nil, nil
	}
	return NoopCloser{Reader: bytes.NewReader(*entry.contents)}, nil
}

func (m *memoryDocumentStore) ContentsSize() int {
	var size int32
	m.mutex.ReadLock(func() {
		size = atomic.LoadInt32(m.contentsSize)
	})
	return int(size)
}

func (m memoryDocumentStore) Size() int {
	var size int
	m.mutex.ReadLock(func() {
		size = len(m.entries)
	})
	return size
}

func (m memoryDocumentStore) Get(hash model.Hash) (*StoredDocument, error) {
	var result *StoredDocument
	var err error
	m.mutex.ReadLock(func() {
		if entry := m.documentHashIndex[hash.String()]; entry != nil {
			result = &StoredDocument{Document: entry.Document.Clone(), HasContents: entry.contents != nil}
		}
	})
	return result, err
}

func (m *memoryDocumentStore) Add(document model.Document) (model.Hash, error) {
	var err error
	var consistencyHash = model.EmptyHash()
	m.mutex.WriteLock(func() {
		if m.documentHashIndex[document.Hash.String()] != nil {
			err = fmt.Errorf("document already exists: %s", document.Hash)
			return
		}
		newEntry := &memoryEntry{Document: document}
		m.entries = append(m.entries, newEntry)
		m.documentHashIndex[newEntry.Hash.String()] = newEntry
		// TODO: Isn't there a faster way to keep it sorted (or not sort it at all?)
		sort.Slice(m.entries, func(i, j int) bool {
			if m.entries[i].Timestamp == m.entries[j].Timestamp {
				return m.entries[i].Timestamp.Before(m.entries[j].Timestamp)
			} else {
				return m.entries[i].Hash.String() < m.entries[j].Hash.String()
			}
		})
		// Calc last consistency hash
		// TODO: Test this
		// TODO: Make this smarter (retain unchanged consistency hashes)
		m.consistencyHashIndex = make(map[string]*memoryEntry, len(m.entries))
		documents := make([]model.Document, len(m.entries))
		var i = 0
		prevHash := model.EmptyHash()
		consistencyHashes := make([]string, len(m.entries))
		for i = 0; i < len(m.entries); i++ {
			documents[i] = m.entries[i].Document
			if i == 0 {
				copy(prevHash, m.entries[i].Hash)
			} else {
				prevHash = model.MakeConsistencyHash(prevHash, m.entries[i].Hash)
			}
			consistencyHashes[i] = prevHash.String()
			m.consistencyHashIndex[prevHash.String()] = m.entries[i]
		}
		m.lastConsistencyHash.Store(prevHash)
		consistencyHash = prevHash
	})
	return consistencyHash, err
}

func (m memoryDocumentStore) LastConsistencyHash() model.Hash {
	var hash model.Hash
	m.mutex.ReadLock(func() {
		hash = m.lastConsistencyHash.Load().(model.Hash).Clone()
	})
	return hash
}

// NoopCloser implements io.ReadCloser with a No-Operation, intended for returning byte slices.
type NoopCloser struct {
	io.Reader
	io.Closer
}

func (NoopCloser) Close() error { return nil }
