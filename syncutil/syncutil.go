// Package syncutil provides utilities for the sync package.
package syncutil

import (
	"sync"
)

// MapFor is a typed wrapper around [sync.Map].
type MapFor[K any, V any] struct {
	m sync.Map
}

// Clear is a wrapper around [sync.Map.Clear].
func (m *MapFor[K, V]) Clear() {
	m.m.Clear()
}

// CompareAndDelete is a wrapper around [sync.Map.CompareAndDelete].
func (m *MapFor[K, V]) CompareAndDelete(key K, old V) (deleted bool) {
	return m.m.CompareAndDelete(key, old)
}

// CompareAndSwap is a wrapper around [sync.Map.CompareAndSwap].
func (m *MapFor[K, V]) CompareAndSwap(key K, oldValue, newValue V) (swapped bool) {
	return m.m.CompareAndSwap(key, oldValue, newValue)
}

// Delete is a wrapper around [sync.Map.Delete].
func (m *MapFor[K, V]) Delete(key K) {
	m.m.Delete(key)
}

// Load is a wrapper around [sync.Map.Load].
func (m *MapFor[K, V]) Load(key K) (value V, ok bool) {
	vi, ok := m.m.Load(key)
	value, _ = vi.(V)
	return value, ok
}

// LoadAndDelete is a wrapper around [sync.Map.LoadAndDelete].
func (m *MapFor[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	vi, loaded := m.m.LoadAndDelete(key)
	value, _ = vi.(V)
	return value, loaded
}

// LoadOrStore is a wrapper around [sync.Map.LoadOrStore].
func (m *MapFor[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	vi, loaded := m.m.LoadOrStore(key, value)
	actual, _ = vi.(V)
	return actual, loaded
}

// Range is a wrapper around [sync.Map.Range].
func (m *MapFor[K, V]) Range(f func(key K, value V) bool) {
	m.m.Range(func(ki, vi any) bool {
		key := ki.(K)   //nolint:forcetypeassert // The map is typed.
		value := vi.(V) //nolint:forcetypeassert // The map is typed.
		return f(key, value)
	})
}

// Store is a wrapper around [sync.Map.Store].
func (m *MapFor[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
}

// Swap is a wrapper around [sync.Map.LoadOrStore].
func (m *MapFor[K, V]) Swap(key K, value V) (previous V, loaded bool) {
	vi, loaded := m.m.LoadOrStore(key, value)
	previous, _ = vi.(V)
	return previous, loaded
}

// PoolFor is a typed wrapper around [sync.Pool].
type PoolFor[T any] struct {
	p   sync.Pool
	New func() *T
}

// Get is a wrapper around [sync.Pool.Get].
func (p *PoolFor[T]) Get() *T {
	vi := p.p.Get()
	if vi != nil {
		return vi.(*T) //nolint:forcetypeassert // The pool is typed.
	}
	if p.New != nil {
		return p.New()
	}
	return nil
}

// Put is a wrapper around [sync.Pool.Put].
func (p *PoolFor[T]) Put(v *T) {
	p.p.Put(v)
}
