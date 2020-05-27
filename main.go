/*
 * Nuts registry
 * Copyright (C) 2019. Nuts community
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

package main

import (
	"github.com/nuts-foundation/nuts-network/cmd"
	"os"
	"strings"
)

func main() {
	e := "NUTS_IDENTITY=urn:oid:1.3.6.1.4.1.54851.4:00000003 NUTS_ADDRESS=localhost:1324 NUTS_PUBLICADDR=localhost:5553 NUTS_GRPCADDR=:5553 NUTS_BOOTSTRAPNODES=localhost:5555"
	for _, keyValue := range strings.Split(e, " ") {
		parts := strings.Split(keyValue, "=")
		os.Setenv(parts[0], parts[1])
	}
	cmd.Execute()
}
