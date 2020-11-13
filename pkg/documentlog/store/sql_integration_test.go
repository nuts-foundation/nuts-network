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
	"bytes"
	"crypto/sha1"
	"fmt"
	logging "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"sort"
	"sync"
	"testing"
	"time"
)

var documentContents = []byte("foobar")
var contentsHash = sha1.Sum(documentContents)

func Test_sqlDocumentStore_AddAndGet(t *testing.T) {
	docStore := new(sqlDocumentStore)
	t.Run("ok - add then get", func(t *testing.T) {
		defer createDatabase(docStore)()
		expected := getDocument()
		consistencyHash, err := docStore.Add(expected)
		if !assert.NoError(t, err) {
			return
		}
		err = docStore.WriteContents(expected.Hash, bytes.NewReader(documentContents))
		if !assert.NoError(t, err) {
			return
		}
		assert.True(t, consistencyHash.Equals(expected.Hash)) // First consistency hash should match document hash

		actual, err := docStore.Get(expected.Hash)
		if !assert.NoError(t, err) {
			return
		}
		assert.True(t, actual.HasContents)
		assert.Equal(t, expected, actual.Document)
	})
	t.Run("ok - concurrency, verify consistency hashes", func(t *testing.T) {
		defer createDatabase(docStore)()
		goRoutines := 20
		docsPerGoroutine := 10
		wg := sync.WaitGroup{}
		wg.Add(goRoutines)
		for x := 0; x < goRoutines/2; x++ {
			fn := func(num int, contents string) {
				for i := 0; i < docsPerGoroutine; i++ {
					ts := int64(num*1000 + i*10)
					_, err := docStore.Add(model.Document{
						Hash:      toHash(fmt.Sprintf("%s%038d", contents, ts)),
						Type:      "doc",
						Timestamp: time.Unix(ts, 0).UTC(),
					})
					if !assert.NoError(t, err) {
						return
					}
				}
				wg.Done()
			}
			go fn(x, "A0")
			// Spawn a second goroutine so we have duplicate timestamps
			go fn(x, "0B")
		}
		wg.Wait()

		printAllDocuments(docStore)
		documents, _ := docStore.GetAll()
		assert.Len(t, documents, docsPerGoroutine*goRoutines)
		prevHash := model.EmptyHash()
		for i, document := range documents {
			prevHash = model.MakeConsistencyHash(prevHash, document.Hash)
			if !assert.Equal(t, prevHash.String(), document.ConsistencyHash.String(), "expected consistency hash for %d (index: %d) to be correct", document.Timestamp.UnixNano(), i) {
				return
			}
		}
	})
	t.Run("error - document already exists", func(t *testing.T) {
		defer createDatabase(docStore)()
		expected := getDocument()
		_, err := docStore.Add(expected)
		if !assert.NoError(t, err) {
			return
		}
		_, err = docStore.Add(expected)
		assert.EqualError(t, err, "UNIQUE constraint failed: document.hash")
	})
}

func Test_sqlDocumentStore_Get(t *testing.T) {
	docStore := new(sqlDocumentStore)
	t.Run("ok - not found", func(t *testing.T) {
		defer createDatabase(docStore)()
		document, err := docStore.Get(model.EmptyHash())
		if !assert.NoError(t, err) {
			return
		}
		assert.Nil(t, document)
	})
}

func Test_sqlDocumentStore_GetByConsistencyHash(t *testing.T) {
	docStore := new(sqlDocumentStore)
	t.Run("ok", func(t *testing.T) {
		defer createDatabase(docStore)()
		h1, _ := docStore.Add(getDocument())
		h2, _ := docStore.Add(getDocument())
		h3, _ := docStore.Add(getDocument())
		d1, _ := docStore.GetByConsistencyHash(h1)
		d2, _ := docStore.GetByConsistencyHash(h2)
		d3, _ := docStore.GetByConsistencyHash(h3)
		assert.Equal(t, h1.String(), d1.ConsistencyHash.String())
		assert.Equal(t, h2.String(), d2.ConsistencyHash.String())
		assert.Equal(t, h3.String(), d3.ConsistencyHash.String())
	})
	t.Run("ok - not found", func(t *testing.T) {
		defer createDatabase(docStore)()
		document, err := docStore.GetByConsistencyHash(model.EmptyHash())
		if !assert.NoError(t, err) {
			return
		}
		assert.Nil(t, document)
	})
}

func Test_sqlDocumentStore_GetAll(t *testing.T) {
	docStore := new(sqlDocumentStore)
	t.Run("ok - empty", func(t *testing.T) {
		defer createDatabase(docStore)()
		docs, err := docStore.GetAll()
		assert.NoError(t, err)
		assert.Len(t, docs, 0)
	})
	t.Run("ok", func(t *testing.T) {
		defer createDatabase(docStore)()
		docStore.Add(getDocument())
		docStore.Add(getDocument())
		docs, err := docStore.GetAll()
		assert.NoError(t, err)
		assert.Len(t, docs, 2)
		for _, doc := range docs {
			assert.NotEmpty(t, doc.Type)
			assert.NotEmpty(t, doc.Timestamp)
			assert.False(t, doc.HasContents)
		}
	})
	t.Run("ok - verify sort order", func(t *testing.T) {
		defer createDatabase(docStore)()
		input := map[string]time.Time{
			"0100000000000000000000000000000000000001": time.Unix(10, 0),
			"0200000000000000000000000000000000000020": time.Unix(10, 0),
			"0400000000000000000000000000000000004000": time.Unix(100, 0),
			"0300000000000000000000000000000000000300": time.Unix(10, 0),
			"0500000000000000000000000000000000050000": time.Unix(5, 0),
		}
		for h, ts := range input {
			_, err := docStore.Add(model.Document{
				Hash:      toHash(h),
				Timestamp: ts,
			})
			if !assert.NoError(t, err, "inserting hash %s", h) {
				printAllDocuments(docStore)
				return
			}
		}
		printAllDocuments(docStore)
		docs, err := docStore.GetAll()
		if !assert.Len(t, docs, 5) {
			return
		}
		assert.NoError(t, err)
		assert.Equal(t, "0500000000000000000000000000000000050000", docs[0].Hash.String())
		assert.Equal(t, "0100000000000000000000000000000000000001", docs[1].Hash.String())
		assert.Equal(t, "0200000000000000000000000000000000000020", docs[2].Hash.String())
		assert.Equal(t, "0300000000000000000000000000000000000300", docs[3].Hash.String())
		assert.Equal(t, "0400000000000000000000000000000000004000", docs[4].Hash.String())
	})
}

func Test_sqlDocumentStore_findPredecessors(t *testing.T) {
	docStore := new(sqlDocumentStore)
	defer createDatabase(docStore)()
	kwik := model.Document{
		Hash:      toHash("0200000000000000000000000000000000000000"),
		Timestamp: time.Unix(10, 0),
	}
	kwek := model.Document{
		Hash:      toHash("0100000000000000000000000000000000000000"),
		Timestamp: time.Unix(10, 0),
	}
	kwak := model.Document{
		Hash:      toHash("0300000000000000000000000000000000000000"),
		Timestamp: time.Unix(00, 0),
	}
	docStore.Add(kwik)
	docStore.Add(kwek)
	docStore.Add(kwak)
	printAllDocuments(docStore)
	// Expected order: kwak, kwek, kwik
	pred, err := docStore.findPredecessors(docStore.db, kwak)
	assert.NoError(t, err)
	assert.Nil(t, pred)

	pred, err = docStore.findPredecessors(docStore.db, kwek)
	assert.NoError(t, err)
	assert.Equal(t, kwak.Hash.String(), pred[0].Hash.String())
}

func Test_sqlDocumentStore_WriteAndReadContents(t *testing.T) {
	docStore := new(sqlDocumentStore)
	t.Run("ok - write then read", func(t *testing.T) {
		defer createDatabase(docStore)()
		expected := getDocument()
		docStore.Add(expected)
		err := docStore.WriteContents(expected.Hash, bytes.NewReader(documentContents))
		assert.NoError(t, err)

		buf := new(bytes.Buffer)
		contents, err := docStore.ReadContents(expected.Hash)
		if !assert.NoError(t, err) {
			return
		}
		_, err = buf.ReadFrom(contents)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, documentContents, buf.Bytes())
	})
}

func Test_sqlDocumentStore_updateContentsHashes(t *testing.T) {
	docStore := new(sqlDocumentStore)
	defer createDatabase(docStore)()
	document := getDocument()
	docStore.Add(document)
	docStore.WriteContents(document.Hash, bytes.NewReader(documentContents))
	// Simulate old record without contents_hash
	result := docStore.db.Exec("UPDATE document SET contents_hash = NULL")
	if !assert.NoError(t, result.Error) {
		return
	}
	if !assert.Equal(t, int64(1), result.RowsAffected) {
		return
	}
	// Verify we can't find the document via its content hash
	actual, _ := docStore.FindByContentsHash(contentsHash)
	assert.Empty(t, actual)
	// Now update missing content hashes
	err := docStore.updateContentsHashes()
	if !assert.NoError(t, err) {
		return
	}
	// Assert we can find the document via its content hash
	actual, _ = docStore.FindByContentsHash(contentsHash)
	assert.Len(t, actual, 1)
}

func Test_sqlDocumentStore_FindByContentsHash(t *testing.T) {
	docStore := new(sqlDocumentStore)
	t.Run("ok - no results", func(t *testing.T) {
		defer createDatabase(docStore)()
		documents, err := docStore.FindByContentsHash(sha1.Sum([]byte("test")))
		if !assert.NoError(t, err) {
			return
		}
		assert.Empty(t, documents)
	})
	t.Run("ok - 1 result", func(t *testing.T) {
		defer createDatabase(docStore)()
		expected := getDocument()
		docStore.Add(expected)
		err := docStore.WriteContents(expected.Hash, bytes.NewReader(documentContents))
		documents, err := docStore.FindByContentsHash(contentsHash)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, documents, 1)
	})
	t.Run("ok - 2 results", func(t *testing.T) {
		defer createDatabase(docStore)()
		// Add 4 documents to the store: 2 with matching contents, 1 with different content and 1 without.
		expected1 := getDocument()
		docStore.Add(expected1)
		docStore.WriteContents(expected1.Hash, bytes.NewReader(documentContents))
		expected2 := getDocument()
		docStore.Add(expected2)
		docStore.WriteContents(expected2.Hash, bytes.NewReader(documentContents))
		otherContents := getDocument()
		otherContents.Hash = model.CalculateDocumentHash(otherContents.Type, otherContents.Timestamp, []byte("otherContents"))
		docStore.Add(otherContents)
		docStore.WriteContents(otherContents.Hash, bytes.NewReader([]byte("otherContents")))
		noContent := getDocument()
		noContent.Hash = model.CalculateDocumentHash(noContent.Type, noContent.Timestamp, []byte("doc2"))
		docStore.Add(noContent)

		documents, err := docStore.FindByContentsHash(contentsHash)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, documents, 2)
	})
}

func Test_sqlDocumentStore_ReadContents(t *testing.T) {
	docStore := new(sqlDocumentStore)
	t.Run("error - document doesn't exist", func(t *testing.T) {
		defer createDatabase(docStore)()
		contents, err := docStore.ReadContents(model.SliceToHash([]byte{1, 2, 3}))
		assert.Error(t, err)
		assert.Nil(t, contents)
	})
	t.Run("ok - document has no no contents", func(t *testing.T) {
		defer createDatabase(docStore)()
		expected := getDocument()
		docStore.Add(expected)
		contents, err := docStore.ReadContents(expected.Hash)
		assert.NoError(t, err)
		assert.Nil(t, contents)
	})
}

func Test_sqlDocumentStore_Size(t *testing.T) {
	docStore := new(sqlDocumentStore)
	t.Run("ok", func(t *testing.T) {
		defer createDatabase(docStore)()
		expected := getDocument()
		docStore.Add(expected)
		assert.Equal(t, 1, docStore.Size())
	})
	t.Run("ok - empty", func(t *testing.T) {
		defer createDatabase(docStore)()
		assert.Equal(t, 0, docStore.Size())
	})
}

func Test_sqlDocumentStore_ContentsSize(t *testing.T) {
	docStore := new(sqlDocumentStore)
	t.Run("ok", func(t *testing.T) {
		defer createDatabase(docStore)()
		document := getDocument()
		docStore.Add(document)
		docStore.WriteContents(document.Hash, bytes.NewReader(documentContents))
		docStore.Add(getDocument())
		assert.Equal(t, len(documentContents), docStore.ContentsSize())
	})
	t.Run("ok - no documents", func(t *testing.T) {
		defer createDatabase(docStore)()
		assert.Equal(t, 0, docStore.Size())
	})
	t.Run("ok - no documents contents ", func(t *testing.T) {
		defer createDatabase(docStore)()
		docStore.Add(getDocument())
		assert.Equal(t, 0, docStore.ContentsSize())
	})
}

func Test_sqlDocumentStore_LastConsistencyHash(t *testing.T) {
	docStore := new(sqlDocumentStore)
	t.Run("ok", func(t *testing.T) {
		defer createDatabase(docStore)()
		document := getDocument()
		docStore.Add(document)
		if !assert.True(t, document.Hash.Equals(docStore.LastConsistencyHash())) {
			return
		}
		consistencyHash, _ := docStore.Add(getDocument())
		if !assert.False(t, document.Hash.Equals(docStore.LastConsistencyHash())) {
			return
		}
		if !assert.True(t, consistencyHash.Equals(docStore.LastConsistencyHash())) {
			return
		}
	})
	t.Run("ok - multiple documents with same timestamp", func(t *testing.T) {
		defer createDatabase(docStore)()
		doc1 := model.Document{Hash: toHash("0100000000000000000000000000000000000000"), Timestamp: time.Unix(10, 0)}
		docStore.Add(doc1)
		doc2 := model.Document{Hash: toHash("0200000000000000000000000000000000000000"), Timestamp: time.Unix(10, 0)}
		docStore.Add(doc2)

		printAllDocuments(docStore)

		docs := []model.Document{doc1, doc2}
		sortDocuments(docs)
		expected := model.MakeConsistencyHash(docs[0].Hash, docs[1].Hash)
		assert.Equal(t, expected.String(), docStore.LastConsistencyHash().String())
	})
	t.Run("ok - no documents", func(t *testing.T) {
		defer createDatabase(docStore)()
		assert.True(t, model.EmptyHash().Equals(docStore.LastConsistencyHash()))
	})
}

func sortDocuments(docs []model.Document) {
	sort.Slice(docs, func(i, j int) bool {
		d1 := docs[i]
		d2 := docs[j]
		if d1.Timestamp.Equal(d2.Timestamp) {
			return d1.Hash.String() < d2.Hash.String()
		}
		return d1.Timestamp.Before(d2.Timestamp)
	})
}

func getDocument() model.Document {
	document := model.Document{
		Type:      "test",
		Timestamp: time.Now().UTC(),
	}
	document.Hash = model.CalculateDocumentHash(document.Type, document.Timestamp, documentContents)
	return document
}

func createDatabase(target *sqlDocumentStore) func() {
	logging.Log().Level = logrus.DebugLevel
	store, err := CreateDocumentStore(":memory:")
	if err != nil {
		panic(err)
	}
	*target = *store.(*sqlDocumentStore)
	return func() {
		if err := target.db.Close(); err != nil {
			panic(err)
		}
	}
}

func toHash(input string) model.Hash {
	hash, err := model.ParseHash(input)
	if err != nil {
		panic(err)
	}
	return hash
}

func printAllDocuments(docStore *sqlDocumentStore) {
	documents, _ := docStore.GetAll()
	for i, document := range documents {
		fmt.Printf("%d  %s     %s    %d\n", i, document.Hash, document.ConsistencyHash, document.Timestamp.UTC().UnixNano())
	}
}
