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

package api

import (
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDocument_fromModel(t *testing.T) {
	expected := model.Document{
		Hash:      test.StringToHash("37076f2cbe014a109d79b61ae9c7191f4cc57afc"),
		Type:      "test",
		Timestamp: time.Now(),
	}
	actual := Document{}
	actual.fromModel(expected)
	assert.Equal(t, actual.Type, expected.Type)
	assert.Equal(t, actual.Hash, expected.Hash.String())
	assert.Equal(t, actual.Timestamp, expected.Timestamp.UnixNano())
}

func TestDocument_toModel(t *testing.T) {
	expected := model.Document{
		Hash:      test.StringToHash("37076f2cbe014a109d79b61ae9c7191f4cc57afc"),
		Type:      "test",
		Timestamp: time.Unix(0, 1591255458635412720),
	}
	actual := Document{
		Hash:      "37076f2cbe014a109d79b61ae9c7191f4cc57afc",
		Timestamp: 1591255458635412720,
		Type:      "test",
	}
	assert.Equal(t, expected, actual.toModel())
}
