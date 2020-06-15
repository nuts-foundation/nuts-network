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

package test

import (
	"github.com/golang/mock/gomock"
	"github.com/nuts-foundation/nuts-network/pkg/model"
)

func EqHash(hash model.Hash) gomock.Matcher {
	return &hashMatcher{expected: hash}
}

type hashMatcher struct {
	expected model.Hash
}

func (h hashMatcher) Matches(x interface{}) bool {
	if actual, ok := x.(model.Hash); !ok {
		return false
	} else {
		return actual.Equals(h.expected)
	}
}

func (h hashMatcher) String() string {
	return "Hash matches: " + h.expected.String()
}

func StringToHash(input string) model.Hash {
	h, _ := model.ParseHash(input)
	return h
}
