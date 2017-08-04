package gu_sync

import (
	"sync"
)

type (
	Sequenced sync.RWMutex
)

func (s *Sequenced) ReadOnly(action func()) {
	sm := (*sync.RWMutex)(s)
	sm.RLock()
	defer sm.RUnlock()
	action()
}

func (s *Sequenced) ReadWrite(action func()) {
	sm := (*sync.RWMutex)(s)
	sm.Lock()
	defer sm.Unlock()
	action()
}

func (s *Sequenced) RWMutex() *sync.RWMutex {
	return (*sync.RWMutex)(s)
}
