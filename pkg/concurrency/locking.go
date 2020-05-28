package concurrency

import "sync"

type SaferRWMutex struct {
	mutex sync.RWMutex
}

func (m *SaferRWMutex) WriteLock(f func()) {
	m.mutex.Lock()
	f()
	m.mutex.Unlock()
}

func (m *SaferRWMutex) ReadLock(f func()) {
	m.mutex.RLock()
	f()
	m.mutex.RUnlock()
}