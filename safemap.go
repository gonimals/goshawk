package main

import "sync"

type SafeMap[K comparable, V any] struct {
	sync.RWMutex
	data map[K]V
}

func (sm *SafeMap[K, V]) Get(key K) V {
	sm.RLock()
	defer sm.RUnlock()
	return sm.data[key]
}

func (sm *SafeMap[K, V]) Set(key K, val V) {
	sm.Lock()
	defer sm.Unlock()
	sm.data[key] = val
}

func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	return &SafeMap[K, V]{
		data: make(map[K]V),
	}
}
