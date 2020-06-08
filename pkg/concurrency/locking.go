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

package concurrency

import (
	log "github.com/nuts-foundation/nuts-network/logging"
	"os"
	"sync"
	"time"
)

type SaferRWMutex interface {
	WriteLock(f func())
	ReadLock(f func())
}

func NewSaferRWMutex(name string) SaferRWMutex {
	if value, ok := os.LookupEnv("NUTS_DEBUG_LOCKS"); ok && value == "true" {
		return &saferRWMutex{debug: true, name: name}
	}
	return &saferRWMutex{}
}

type saferRWMutex struct {
	mutex sync.RWMutex
	debug bool
	name  string
}

func (m *saferRWMutex) WriteLock(f func()) {
	var t time.Time
	if m.debug {
		log.Log().Infof("Acquiring lock (name=%s)", m.name)
		t = time.Now()
	}
	m.mutex.Lock()
	if m.debug {
		log.Log().Infof("Lock acquired in %d ms (name=%s)", (time.Now().UnixNano()-t.UnixNano())/1000, m.name)
		t = time.Now()
	}
	f()
	m.mutex.Unlock()
	if m.debug {
		log.Log().Infof("Lock released after %d ms (name=%s)", (time.Now().UnixNano()-t.UnixNano())/1000, m.name)
	}
}

func (m *saferRWMutex) ReadLock(f func()) {
	m.mutex.RLock()
	f()
	m.mutex.RUnlock()
}
