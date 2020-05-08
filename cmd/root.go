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
//cmd contains the helper command for running the registry engine as standalone command/server
package cmd

import (
	cfg "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-network/engine"
	"github.com/sirupsen/logrus"
	"os"
)

var e = engine.NewNetworkEngine()
var rootCmd = e.Cmd

func Execute() {
	// TODO: Remove this
	os.Setenv("NUTS_IDENTITY", "urn:oid:1.3.6.1.4.1.54851.4:00000001")

	c := cfg.NewNutsGlobalConfig()
	c.IgnoredPrefixes = append(c.IgnoredPrefixes, e.ConfigKey)
	c.RegisterFlags(rootCmd, e)
	if err := c.Load(rootCmd); err != nil {
		panic(err)
	}

	cfg.RegisterEngine(e)
	c.PrintConfig(logrus.StandardLogger())

	if err := c.InjectIntoEngine(e); err != nil {
		panic(err)
	}

	if err := e.Configure(); err != nil {
		panic(err)
	}

	if err := e.Start(); err != nil {
		panic(err)
	}

	defer e.Shutdown()

	rootCmd.Execute()

	println("EXIT")
}
