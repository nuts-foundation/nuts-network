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

package engine

import (
	core "github.com/nuts-foundation/nuts-go-core"
	logging "github.com/nuts-foundation/nuts-network/logging"
	pkg "github.com/nuts-foundation/nuts-network/pkg"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// NewNetworkEngine returns the core definition for the network
func NewNetworkEngine() *core.Engine {
	engine := pkg.NetworkInstance()
	return &core.Engine{
		Cmd:       cmd(),
		Configure: engine.Configure,
		Config:    &engine.Config,
		ConfigKey: "network",
		FlagSet:   flagSet(),
		Name:      pkg.ModuleName,
		Start:     engine.Start,
		Shutdown:  engine.Shutdown,
	}
}

func flagSet() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("network", pflag.ContinueOnError)
	return flagSet
}

func cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network",
		Short: "network commands",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number of the Nuts network",
		Run: func(cmd *cobra.Command, args []string) {
			logging.Log().Errorf("version 0.0.0")
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "server",
		Short: "Run standalone api server",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO
		},
	})
	return cmd
}