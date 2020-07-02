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

package engine

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-network/api"
	"github.com/nuts-foundation/nuts-network/client"
	logging "github.com/nuts-foundation/nuts-network/logging"
	"github.com/nuts-foundation/nuts-network/pkg"
	"github.com/nuts-foundation/nuts-network/pkg/documentlog"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"io"
	"os"
	"os/signal"
	"strings"
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
		Routes: func(router core.EchoRouter) {
			api.RegisterHandlers(router, &api.ApiWrapper{Service: pkg.NetworkInstance()})
		},
		Name:     pkg.ModuleName,
		Start:    engine.Start,
		Shutdown: engine.Shutdown,
	}
}

func Cmd() *cobra.Command {
	return cmdWithClient(func() pkg.NetworkClient {
		return client.NewNetworkClient()
	})
}

func flagSet() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("network", pflag.ContinueOnError)
	flagSet.String("grpcAddr", ":5555", "Local address for gRPC to listen on.")
	flagSet.String("publicAddr", "", "Public address (of this node) other nodes can use to connect to it. If set, it is registered on the nodelist.")
	flagSet.String("bootstrapNodes", "", "Space-separated list of bootstrap nodes (`<host>:<port>`) which the node initially connect to.")
	flagSet.String("nodeID", "", "Instance ID of this node under which the public address is registered on the nodelist. If not set, the Nuts node's identity will be used.")
	flagSet.String("mode", "", "server or client, when client it uses the HttpClient")
	flagSet.String("address", "", "Interface and port for http server to bind to, defaults to global Nuts address.")
	flagSet.String("certFile", "", "PEM file containing the certificate this node will identify itself with to other nodes. If not set, the Nuts node will attempt to load a TLS certificate from the crypto module.")
	flagSet.String("certKeyFile", "", "PEM file containing the key belonging to this node's certificate. If not set, the Nuts node will attempt to load a TLS certificate from the crypto module.")
	return flagSet
}

func cmdWithClient(clientCreator func() pkg.NetworkClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network",
		Short: "network commands",
	}
	cmd.AddCommand(versionCommand())
	cmd.AddCommand(serverCommand())
	cmd.AddCommand(listCommand(clientCreator))
	cmd.AddCommand(getCommand(clientCreator))
	cmd.AddCommand(contentsCommand(clientCreator))
	return cmd
}

func contentsCommand(clientCreator func() pkg.NetworkClient) *cobra.Command {
	return &cobra.Command{
		Use:   "contents [hash]",
		Short: "Retrieves the contents of a document from the network",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hash, err := model.ParseHash(args[0])
			if err != nil {
				return err
			}
			reader, err := clientCreator().GetDocumentContents(hash)
			if err != nil {
				return err
			}
			if reader == nil {
				logging.Log().Warnf("Document or contents not found: %s", hash)
				return nil
			}
			defer reader.Close()
			buf := new(strings.Builder)
			_, err = io.Copy(buf, reader)
			if err != nil {
				logging.Log().Warnf("Unable to read contents: %v", err)
				return nil
			}
			println(buf.String())
			return nil
		},
	}
}

func getCommand(clientCreator func() pkg.NetworkClient) *cobra.Command {
	return &cobra.Command{
		Use:   "get [hash]",
		Short: "Gets a document from the network",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hash, err := model.ParseHash(args[0])
			if err != nil {
				return err
			}
			instance := clientCreator()
			document, err := instance.GetDocument(hash)
			if err != nil {
				return err
			}
			if document == nil {
				logging.Log().Warnf("Document not found: %s", hash)
				return nil
			}
			logging.Log().Infof("Document %s:\n  Type: %s\n  Timestamp: %s", document.Hash, document.Type, document.Timestamp)
			contents, err := instance.GetDocumentContents(hash)
			if err != nil {
				if errors.Is(err, documentlog.ErrMissingDocumentContents) {
					logging.Log().Info("No contents for the given document on the local node.")
					err = nil
				}
				return err
			}
			buf := new(bytes.Buffer)
			if _, err = buf.ReadFrom(contents); err != nil {
				return err
			}
			println()
			println(string(buf.Bytes()))
			return nil
		},
	}
}

func listCommand(clientCreator func() pkg.NetworkClient) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Lists the documents on the network",
		RunE: func(cmd *cobra.Command, args []string) error {
			documents, err := clientCreator().ListDocuments()
			if err != nil {
				return err
			}
			const format = "%-40s %-40s %-20s\n"
			fmt.Printf(format, "Hash", "Timestamp", "Type")
			for _, document := range documents {
				fmt.Printf(format, document.Hash, document.Timestamp, document.Type)
			}
			return nil
		},
	}
}

func serverCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Run standalone api server",
		Run: func(cmd *cobra.Command, args []string) {
			echoInstance := echo.New()
			echoInstance.HideBanner = true
			echoInstance.Use(middleware.Logger())
			core.NewStatusEngine().Routes(echoInstance)
			api.RegisterHandlers(echoInstance, &api.ApiWrapper{Service: pkg.NetworkInstance()})

			// todo move to nuts-go-core
			sigc := make(chan os.Signal, 1)
			signal.Notify(sigc, os.Interrupt, os.Kill)

			recoverFromEcho := func() {
				defer func() {
					recover()
				}()
				echoInstance.Start(core.NutsConfig().ServerAddress())
			}

			go recoverFromEcho()
			<-sigc
		},
	}
}

func versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of the Nuts network",
		Run: func(cmd *cobra.Command, args []string) {
			logging.Log().Info("version 0.0.0")
		},
	}
}
