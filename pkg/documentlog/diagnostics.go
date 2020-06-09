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
	"strings"
	"sync/atomic"
)

type LastConsistencyHashDiagnostic struct {
	Hash string
}

func (d LastConsistencyHashDiagnostic) Name() string {
	return "[DocumentLog] Last consistency hash"
}

func (d LastConsistencyHashDiagnostic) String() string {
	return d.Hash
}

type ConsistencyHashListDiagnostic struct {
	Hashes []string
}

func (d ConsistencyHashListDiagnostic) Name() string {
	return "[DocumentLog] Ordered consistency hash list"
}

func (d ConsistencyHashListDiagnostic) String() string {
	var result = make([]string, len(d.Hashes))
	for i, hash := range d.Hashes {
		result[i] = fmt.Sprintf("%d->%s", i, hash)
	}
	return strings.Join(result, " ")
}

type NumberOfDocumentsDiagnostic struct {
	NumberOfDocuments int
}

func (d NumberOfDocumentsDiagnostic) Name() string {
	return "[DocumentLog] Number of documents"
}

func (d NumberOfDocumentsDiagnostic) String() string {
	return fmt.Sprintf("%d", d.NumberOfDocuments)
}

type LogSizeDiagnostic struct {
	sizeInBytes int64
}

func (d *LogSizeDiagnostic) add(bytes int64) {
	atomic.AddInt64(&d.sizeInBytes, bytes)
}

func (d LogSizeDiagnostic) Name() string {
	return "[DocumentLog] Stored document size (bytes)"
}

func (d LogSizeDiagnostic) String() string {
	size := atomic.LoadInt64(&d.sizeInBytes)
	return fmt.Sprintf("%d", size)
}
