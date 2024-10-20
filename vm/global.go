package vm

import (
	"sync"
)

type ConcurrencyMap[TKey comparable, TValue any] struct {
	m       map[TKey]TValue
	rwMutex sync.RWMutex
}

func (m *ConcurrencyMap[TKey, TValue]) Get(key TKey) (TValue, bool) {
	m.rwMutex.RLock()
	defer m.rwMutex.RUnlock()
	v, ok := m.m[key]
	return v, ok
}
func (m *ConcurrencyMap[TKey, TValue]) Set(key TKey, value TValue) {
	m.rwMutex.Lock()
	defer m.rwMutex.Unlock()
	m.m[key] = value
}
func (m *ConcurrencyMap[TKey, TValue]) Remove(key TKey) {
	m.rwMutex.Lock()
	defer m.rwMutex.Unlock()
	delete(m.m, key)
}
func (m *ConcurrencyMap[TKey, TValue]) Size() int {
	m.rwMutex.RLock()
	defer m.rwMutex.RUnlock()
	return len(m.m)
}
func (m *ConcurrencyMap[TKey, TValue]) Clear() {
	m.rwMutex.Lock()
	defer m.rwMutex.Unlock()
	m.m = make(map[TKey]TValue)
}
func (m *ConcurrencyMap[TKey, TValue]) Delete(key TKey) {
	m.rwMutex.Lock()
	defer m.rwMutex.Unlock()
	delete(m.m, key)
}
func (m *ConcurrencyMap[TKey, TValue]) Range(rangeFunc func(TKey, TValue) bool) {
	m.rwMutex.RLock()
	defer m.rwMutex.RUnlock()
	for k, v := range m.m {
		rangeFunc(k, v)
	}
}

func NewConcurrencyMap[K comparable, V any](capacity int) *ConcurrencyMap[K, V] {
	return &ConcurrencyMap[K, V]{
		m: make(map[K]V, capacity),
	}
}
