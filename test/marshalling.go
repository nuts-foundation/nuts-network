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
	"bytes"
	"encoding/json"
	"io"
)

type JSONReader struct {
	Input interface{}
	reader io.Reader
}

func (j *JSONReader) Read(p []byte) (n int, err error) {
	if j.reader == nil {
		data, err := json.Marshal(j.Input)
		if err != nil {
			return 0, err
		}
		j.reader = bytes.NewReader(data)
	}
	return j.reader.Read(p)
}

