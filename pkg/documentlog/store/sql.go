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
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	"github.com/jinzhu/gorm"
	logging "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/migrations"
	errors2 "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sync"

	// import needed to enable the sqlite dialect for gorm
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"io"
	"time"
)

const documentTable = "document"

var documentSelectCols = []string{"hash", "type", "timestamp", "consistency_hash", "LENGTH(contents) as size"}

// CreateDocumentStore creates a new DocumentStore given the connection string. The conenction string should be understood
// by the underlying implementation (currently sqlite3).
func CreateDocumentStore(connectionString string) (DocumentStore, error) {
	db, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	if gormDb, err := gorm.Open("sqlite3", db); err != nil {
		return nil, err
	} else {
		if logging.Log().Level >= logrus.DebugLevel {
			gormDb = gormDb.Debug()
		}
		gormDb.SetLogger(logging.Log())
		return newSQLDocumentStore(gormDb)
	}
}

func newSQLDocumentStore(db *gorm.DB) (DocumentStore, error) {
	docStore := &sqlDocumentStore{db: db, writeMutex: &sync.Mutex{}}
	if err := docStore.migrate(); err != nil {
		return nil, err
	}
	return docStore, nil
}

type sqlDocumentStore struct {
	db         *gorm.DB
	writeMutex *sync.Mutex
}

type sqlDocument struct {
	Hash            string `gorm:"type:char(40);PRIMARY_KEY"`
	ConsistencyHash string `gorm:"column:consistency_hash;type:char(40)"`
	Type            string `gorm:"type:varchar(100)"`
	Timestamp       int64  `gorm:"type:bigint"`
	Size            int    `gorm:"-"`
	Contents        []byte `gorm:"type:blob"`
}

func (d sqlDocument) descriptor() model.DocumentDescriptor {
	hash, _ := model.ParseHash(d.Hash)
	consistencyHash, _ := model.ParseHash(d.ConsistencyHash)
	return model.DocumentDescriptor{
		HasContents:     d.Size > 0,
		ConsistencyHash: consistencyHash,
		Document: model.Document{
			Hash:      hash,
			Type:      d.Type,
			Timestamp: time.Unix(0, d.Timestamp).UTC(),
		},
	}
}

func (s sqlDocumentStore) Add(document model.Document) (model.Hash, error) {
	// This mutex is ONLY required for sqlite, so when switching to another database system this should probably be removed!
	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()
	consistencyHash := model.EmptyHash()
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Insert
		if err := tx.Table(documentTable).Create(sqlDocument{
			Hash:      document.Hash.String(),
			Type:      document.Type,
			Timestamp: document.Timestamp.UTC().UnixNano(),
		}).Error; err != nil {
			return err
		}
		if result, err := s.updateConsistencyHashes(tx, document); err != nil {
			return err
		} else {
			consistencyHash = result
		}
		return nil
	})
	return consistencyHash, err
}

func (s sqlDocumentStore) updateConsistencyHashes(tx *gorm.DB, startPoint model.Document) (model.Hash, error) {
	// Consistency hashes use the previous record's consistency hash to calculate the consistency hash for the current record.
	// Thus, we need the consistency has from the record before the starting point to calculate the consistency hash for the
	// startPoint record.
	consistencyHash := model.EmptyHash()
	prevRecords, err := s.findPredecessors(tx, startPoint)
	if err != nil {
		return model.EmptyHash(), err
	}
	if len(prevRecords) > 0 {
		consistencyHash = prevRecords[len(prevRecords)-1].ConsistencyHash
	}

	// Find all documents which' consistency hashes need updating
	documents := make([]sqlDocument, 0)
	query := tx.Table(documentTable).
		Select(documentSelectCols).
		Where("timestamp >= ?", startPoint.Timestamp.UTC().UnixNano())
	for _, prevRecord := range prevRecords {
		query = query.Where("hash <> ?", prevRecord.Hash.String())
	}
	query = query.Order("timestamp, hash ASC").Find(&documents)
	if query.Error != nil {
		return model.EmptyHash(), query.Error
	}
	// Recalculate consistency hashes for found records
	logging.Log().Debugf("Updating consistency hashes for %d records", len(documents))
	for _, document := range documents {
		hash, err := model.ParseHash(document.Hash)
		if err != nil {
			return consistencyHash, err
		}
		consistencyHash = model.MakeConsistencyHash(consistencyHash, hash)
		query := tx.Table(documentTable).Where("hash = ?", document.Hash).UpdateColumn("consistency_hash", consistencyHash.String())
		if query.Error != nil {
			return model.EmptyHash(), errors2.Wrapf(query.Error, "unable to update consistency hash to %s for document %s", consistencyHash, document.Hash)
		}
		if query.RowsAffected != 1 {
			return model.EmptyHash(), errors.New("no rows updated")
		}
	}
	return consistencyHash, nil
}
func (s sqlDocumentStore) get(query interface{}, args ...interface{}) (*model.DocumentDescriptor, error) {
	document := sqlDocument{}
	err := s.db.
		Table(documentTable).
		Select(documentSelectCols).
		Where(query, args).
		First(&document).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	result := document.descriptor()
	return &result, nil
}

func (s sqlDocumentStore) Get(hash model.Hash) (*model.DocumentDescriptor, error) {
	return s.get("hash = ?", hash.String())
}

func (s sqlDocumentStore) GetByConsistencyHash(hash model.Hash) (*model.DocumentDescriptor, error) {
	return s.get("consistency_hash = ?", hash.String())
}

func (s sqlDocumentStore) GetAll() ([]model.DocumentDescriptor, error) {
	documents := make([]sqlDocument, 0)
	err := s.db.Table(documentTable).Select(documentSelectCols).Order("timestamp, hash ASC").Find(&documents).Error
	if err != nil {
		return nil, err
	}
	result := make([]model.DocumentDescriptor, len(documents))
	for i, document := range documents {
		result[i] = document.descriptor()
	}
	return result, nil
}

func (s sqlDocumentStore) WriteContents(hash model.Hash, contents io.Reader) error {
	// TODO: I'd rather stream the contents to the blob, instead of reading it as byte buffer which might be problematic
	// with large documents
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(contents); err != nil {
		return err
	}
	query := s.db.Table(documentTable).Where("hash = ?", hash.String()).UpdateColumn("contents", buf.Bytes())
	if query.Error != nil {
		return query.Error
	}
	if query.RowsAffected == 0 {
		return fmt.Errorf("document not found: %s", hash)
	}
	return nil
}

func (s sqlDocumentStore) ReadContents(hash model.Hash) (io.ReadCloser, error) {
	document := sqlDocument{}
	err := s.db.
		Table(documentTable).
		Select("contents").
		Where("hash = ?", hash.String()).
		First(&document).Error
	if err != nil {
		return nil, err
	}
	if document.Contents == nil {
		return nil, nil
	}
	return &NoopCloser{Reader: bytes.NewReader(document.Contents)}, nil
}

func (s sqlDocumentStore) LastConsistencyHash() model.Hash {
	values := make([]interface{}, 0)
	query := s.db.
		Table(documentTable).
		Select("consistency_hash").
		Where("timestamp = ?", s.db.Table(documentTable).Select("MAX(timestamp)").SubQuery()).
		Order("hash ASC").
		Pluck("consistency_hash", &values)
	if query.Error != nil {
		logging.Log().Errorf("Unable to determine last consistency hash: %v", query.Error)
		return model.EmptyHash()
	}
	// We map to interface{} to account for NULL value
	if len(values) == 0 {
		return model.EmptyHash()
	}
	hash, _ := model.ParseHash(values[len(values)-1].(string))
	return hash
}

// findPredecessor finds the document before the specified document (in the chain). If there are multiple
// documents at that (preceding) timestamp, they are all returned. If multiple documents are returned, they
// are in order as they appear in the chain (ordered by hash).
func (s sqlDocumentStore) findPredecessors(tx *gorm.DB, before model.Document) ([]model.DocumentDescriptor, error) {
	documents := make([]sqlDocument, 0)
	ts := before.Timestamp.UTC().UnixNano()
	// Query is in raw form, because I couldn't built it using gorm (missing unions and select-from-subquery).
	err := tx.Raw(`SELECT * FROM (
	  SELECT * FROM document WHERE hash <> ? AND timestamp = ?
	  UNION
	  SELECT * FROM document WHERE TIMESTAMP = (SELECT MAX(timestamp) FROM document WHERE timestamp < ?)
	) ORDER BY timestamp, hash ASC`, before.Hash.String(), ts, ts).Find(&documents).Error
	if err != nil {
		return nil, err
	}
	if len(documents) == 0 {
		// 'before' is the first record
		return nil, nil
	}
	result := make([]model.DocumentDescriptor, 0)
	for _, document := range documents {
		// Account for documents with the same timestamp but a larger hash (they come after the current hash)
		if document.Timestamp == ts && document.Hash > before.Hash.String() {
			continue
		}
		result = append(result, document.descriptor())
	}
	return result, nil
}

func (s sqlDocumentStore) ContentsSize() int {
	values := make([]interface{}, 1)
	query := s.db.Table("stats").Where("key = 'contents-size'").Pluck("CAST(value AS INTEGER)", &values)
	if query.Error != nil {
		logging.Log().Errorf("Unable to determine contents size: %v", query.Error)
		return -1
	}
	// We map to interface{} to account for NULL value
	if values[0] == nil {
		return 0
	}
	return int(values[0].(int64))
}

func (s sqlDocumentStore) Size() int {
	size := 0
	if err := s.db.Table(documentTable).Count(&size).Error; err != nil {
		logging.Log().Errorf("Unable to determine size: %v", err)
		return -1
	}
	return size
}

func (s sqlDocumentStore) migrate() error {
	driver, err := sqlite3.WithInstance(s.db.DB(), &sqlite3.Config{})
	if err != nil {
		return err
	}
	resource := bindata.Resource(migrations.AssetNames(),
		func(name string) ([]byte, error) {
			return migrations.Asset(name)
		})
	source, err := bindata.WithInstance(resource)
	if err != nil {
		return err
	}
	migrator, err := migrate.NewWithInstance("go-bindata", source, "test", driver)
	if err != nil {
		return err
	}
	err = migrator.Up()
	if err != nil && err.Error() != "no change" {
		return err
	}
	return nil
}

// NoopCloser implements io.ReadCloser with a No-Operation, intended for returning byte slices.
type NoopCloser struct {
	io.Reader
	io.Closer
}

func (NoopCloser) Close() error { return nil }
