package util

import (
	"reflect"
	"sync"
)

type SyncMap[K comparable, V any] struct {
	data             map[K]V
	mu               *sync.RWMutex
	comparableValues bool
}

func (sm *SyncMap[K, V]) Get(key K) (val V) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.data[key]
}

func (sm *SyncMap[K, V]) Set(key K, val V) {
	sm.mu.Lock()
	sm.data[key] = val
	sm.mu.Unlock()
}

func (sm *SyncMap[K, V]) CompareAndSwap(key K, oldVal, newVal V) bool {
	if !sm.comparableValues {
		return false
	}
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if any(oldVal) != any(sm.data[key]) {
		return false
	}
	sm.data[key] = newVal
	return true
}

func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{
		data:             map[K]V{},
		mu:               &sync.RWMutex{},
		comparableValues: reflect.TypeFor[V]().Comparable(),
	}
}
