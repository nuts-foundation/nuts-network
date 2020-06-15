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

package client

import (
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-network/api"
	"github.com/nuts-foundation/nuts-network/pkg"
	"github.com/sirupsen/logrus"
	"time"
)

// NewRegistryClient creates a new Local- or RemoteClient for the nuts registry
func NewNetworkClient() pkg.NetworkClient {
	instance := pkg.NetworkInstance()

	if core.NutsConfig().GetEngineMode(instance.Config.Mode) == core.ServerEngineMode {
		if err := instance.Configure(); err != nil {
			logrus.Panic(err)
		}

		return instance
	} else {
		return api.HttpClient{
			ServerAddress: instance.Config.Address,
			Timeout:       30 * time.Second,
		}
	}
}
