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
	"fmt"
	"github.com/golang/mock/gomock"
)

func IsOfType(t string) gomock.Matcher {
	return &typeMatcher{typeName: t}
}

type typeMatcher struct {
	typeName     string
	resolvedType string
}

func (b *typeMatcher) Matches(x interface{}) bool {
	b.resolvedType = fmt.Sprintf("%T", x)
	return b.resolvedType == b.typeName
}

func (b typeMatcher) String() string {
	return "type is " + b.typeName + " (actual: " + b.resolvedType + ")"
}
