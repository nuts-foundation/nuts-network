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
	flagSet.String("grpcAddr", ":5555", "gRPC address to listen on")
	flagSet.String("publicAddr", "", "Public address other nodes can connect to. If empty, the node will not be registered on the nodelist.")
	flagSet.String("bootstrapNodes", "", "Space-separated list of bootstrap node addresses")
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
			logging.Log().Info("version 0.0.0")
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "server",
		Short: "Run standalone api server",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO
			c := make(chan bool, 1)
			<-c
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use: "add",
		Short: "Add a document to the network",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pkg.NetworkInstance().AddDocument([]byte(args[0]))
		},
	})
	return cmd
}
