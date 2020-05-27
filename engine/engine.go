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
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	core "github.com/nuts-foundation/nuts-go-core"
	logging "github.com/nuts-foundation/nuts-network/logging"
	pkg "github.com/nuts-foundation/nuts-network/pkg"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
	"os/signal"
)

// NewNetworkEngine returns the core definition for the network
func NewNetworkEngine() *core.Engine {
	engine := pkg.NetworkInstance()
	return &core.Engine{
		Cmd:         Cmd(),
		Configure:   engine.Configure,
		Config:      &engine.Config,
		ConfigKey:   "network",
		Diagnostics: engine.Diagnostics,
		FlagSet:     flagSet(),
		Name:        pkg.ModuleName,
		Start:       engine.Start,
		Shutdown:    engine.Shutdown,
	}
}

func flagSet() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("network", pflag.ContinueOnError)
	flagSet.String("grpcAddr", ":5555", "gRPC address to listen on")
	flagSet.String("publicAddr", "", "Public address other nodes can connect to. If empty, the node will not be registered on the nodelist.")
	flagSet.String("bootstrapNodes", "", "Space-separated list of bootstrap node addresses")
	flagSet.String("nodeId", "", "ID of this node. If not set, the node's identity will be used.")
	return flagSet
}

func Cmd() *cobra.Command {
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
			echo := echo.New()
			echo.HideBanner = true
			echo.Use(middleware.Logger())
			core.NewStatusEngine().Routes(echo)

			// todo move to nuts-go-core
			sigc := make(chan os.Signal, 1)
			signal.Notify(sigc, os.Interrupt, os.Kill)

			recoverFromEcho := func() {
				defer func() {
					recover()
				}()
				echo.Start(core.NutsConfig().ServerAddress())
			}

			go recoverFromEcho()
			<-sigc
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "add",
		Short: "Add a document to the network",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pkg.NetworkInstance().AddDocument([]byte(args[0]))
		},
	})
	return cmd
}
