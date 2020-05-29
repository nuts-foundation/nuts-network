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
