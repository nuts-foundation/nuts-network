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
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/pkg/proto"
)

type DocumentLog interface {
	proto.HashSource
	// Starts the document log
	Start()
	Stop()
	Subscribe(documentType string) DocumentQueue // TODO: Subscribe is a bad name when returning a blocking queue

	Diagnostics() []core.DiagnosticResult
}

type DocumentQueue interface {
	Get() model.Document
}

type documentQueue struct {
	documentType string
	c            chan model.Document
}

func (q documentQueue) Get() model.Document {
	return <-q.c
}
