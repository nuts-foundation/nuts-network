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

package pkg

import (
	"sync"
)

// ModuleName defines the name of this module
const ModuleName = "Network"

// NetworkClient is the interface to be implemented by any remote or local client
type NetworkClient interface {

}

// NetworkConfig holds the config
type NetworkConfig struct {
}

// Network holds the config and Db reference
type Network struct {
	Config      NetworkConfig
	configOnce  sync.Once
}

var instance *Network
var oneRegistry sync.Once

// NetworkInstance returns the singleton Network
func NetworkInstance() *Network {
	oneRegistry.Do(func() {
		instance = &Network{
		}
	})

	return instance
}

// Configure initializes the db, but only when in server mode
func (r *Network) Configure() error {
	var err error

	r.configOnce.Do(func() {
		// TODO
	})
	return err
}

// Start initiates the network subsystem
func (r *Network) Start() error {
	return nil
}

// Shutdown cleans up any leftover go routines
func (r *Network) Shutdown() error {
	return nil
}