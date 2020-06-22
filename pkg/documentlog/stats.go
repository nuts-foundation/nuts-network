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
	"fmt"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"strings"
)

// LastConsistencyHashStatistic holds the last consistency hash stored on the DocumentLog.
type LastConsistencyHashStatistic struct {
	// Hash is the last consistency hash.
	Hash model.Hash
}

func (d LastConsistencyHashStatistic) Name() string {
	return "[DocumentLog] Last consistency hash"
}

func (d LastConsistencyHashStatistic) String() string {
	return d.Hash.String()
}

// ConsistencyHashListStatistic holds an ordered list of all consistency hashes on the Document Log.
type ConsistencyHashListStatistic struct {
	// Hashes contains all consistency hashes
	Hashes []string
}

func (d ConsistencyHashListStatistic) Name() string {
	return "[DocumentLog] Ordered consistency hash list"
}

func (d ConsistencyHashListStatistic) String() string {
	var result = make([]string, len(d.Hashes))
	for i, hash := range d.Hashes {
		result[i] = fmt.Sprintf("%d->%s", i, hash)
	}
	return strings.Join(result, " ")
}

// NumberOfDocumentsStatistic holds the number of documents stored on the DocumentLog.
type NumberOfDocumentsStatistic struct {
	NumberOfDocuments int
}

func (d NumberOfDocumentsStatistic) Name() string {
	return "[DocumentLog] Number of documents"
}

func (d NumberOfDocumentsStatistic) String() string {
	return fmt.Sprintf("%d", d.NumberOfDocuments)
}

// LogSizeStatistic holds the storage size of all stored document contents (in bytes).
type LogSizeStatistic struct {
	sizeInBytes int
}

func (d LogSizeStatistic) Name() string {
	return "[DocumentLog] Stored document size (bytes)"
}

func (d LogSizeStatistic) String() string {
	return fmt.Sprintf("%d", d.sizeInBytes)
}
