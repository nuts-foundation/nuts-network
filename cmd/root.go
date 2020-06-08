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
//cmd contains the helper command for running the registry engine as standalone command/server
package cmd

import (
	cfg "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-network/engine"
	"github.com/sirupsen/logrus"
)

func Execute() {
	var engineInstance = engine.NewNetworkEngine()
	globalConfig := cfg.NutsConfig()
	globalConfig.IgnoredPrefixes = append(globalConfig.IgnoredPrefixes, engineInstance.ConfigKey)
	globalConfig.RegisterFlags(engineInstance.Cmd, engineInstance)
	if err := globalConfig.Load(engineInstance.Cmd); err != nil {
		panic(err)
	}

	cfg.RegisterEngine(engineInstance)
	globalConfig.PrintConfig(logrus.StandardLogger())

	if err := globalConfig.InjectIntoEngine(engineInstance); err != nil {
		panic(err)
	}

	if err := engineInstance.Configure(); err != nil {
		panic(err)
	}

	if err := engineInstance.Start(); err != nil {
		panic(err)
	}

	defer engineInstance.Shutdown()

	engineInstance.Cmd.Execute()

	println("EXIT")
}
