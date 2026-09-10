package weakutil

import (
	"iter"
	"runtime"
	"sync"
	"weak"

	"github.com/pierrre/go-libs/syncutil"
)

// ValueMap is a map that automatically evicts entries when the value is no longer reachable.
// It is safe for concurrent use.
// The zero value is ready to use.
// If a nil value is set, it is never evicted.
//
// It implements the same methods as [sync.Map].
// An entry whose value has been collected by the garbage collector is treated as absent until its cleanup evicts it.
type ValueMap[K comparable, V any] struct {
	m               syncutil.Map[K, valueMapValue[V]]
	cleanupFunc     func(valueMapCleanupArg[K, V])
	cleanupFuncOnce sync.Once
}

type valueMapValue[T any] struct {
	value   weak.Pointer[T]
	cleanup runtime.Cleanup
}

func (m *ValueMap[K, V]) getCleanupFunc() func(valueMapCleanupArg[K, V]) {
	m.cleanupFuncOnce.Do(func() {
		m.cleanupFunc = m.cleanup
	})
	return m.cleanupFunc
}

func (m *ValueMap[K, V]) newValue(key K, value *V) valueMapValue[V] {
	var mv valueMapValue[V]
	if value != nil {
		mv.value = weak.Make(value)
		mv.cleanup = runtime.AddCleanup(value, m.getCleanupFunc(), valueMapCleanupArg[K, V]{
			key:     key,
			pointer: mv.value,
		})
	}
	return mv
}

type valueMapCleanupArg[K comparable, V any] struct {
	key     K
	pointer weak.Pointer[V]
}

func (m *ValueMap[K, V]) cleanup(mc valueMapCleanupArg[K, V]) {
	mv, ok := m.m.Load(mc.key)
	if ok && mv.value == mc.pointer {
		m.m.CompareAndDelete(mc.key, mv)
	}
}

// Store is like [sync.Map.Store].
func (m *ValueMap[K, V]) Store(key K, value *V) {
	_, _ = m.Swap(key, value)
}

// Load is like [sync.Map.Load].
func (m *ValueMap[K, V]) Load(key K) (value *V, ok bool) {
	mv, ok := m.m.Load(key)
	if !ok {
		return nil, false
	}
	return loadPointer(mv.value)
}

// Delete is like [sync.Map.Delete].
func (m *ValueMap[K, V]) Delete(key K) {
	_, _ = m.LoadAndDelete(key)
}

// Clear is like [sync.Map.Clear].
func (m *ValueMap[K, V]) Clear() {
	for {
		var count int64
		m.m.Range(func(k K, mv valueMapValue[V]) bool {
			count++
			m.m.CompareAndDelete(k, mv)
			mv.cleanup.Stop()
			return true
		})
		if count == 0 {
			return
		}
	}
}

// Swap is like [sync.Map.Swap].
func (m *ValueMap[K, V]) Swap(key K, value *V) (previous *V, loaded bool) {
	previous, loaded = m.Load(key)
	if loaded && previous == value {
		return previous, true
	}
	mv := m.newValue(key, value)
	mv, ok := m.m.Swap(key, mv)
	if ok {
		previous, loaded = loadPointer(mv.value)
		mv.cleanup.Stop()
	}
	return previous, loaded
}

// LoadAndDelete is like [sync.Map.LoadAndDelete].
func (m *ValueMap[K, V]) LoadAndDelete(key K) (value *V, loaded bool) {
	mv, ok := m.m.LoadAndDelete(key)
	if ok {
		value, loaded = loadPointer(mv.value)
		mv.cleanup.Stop()
	}
	return value, loaded
}

// LoadOrStore is like [sync.Map.LoadOrStore].
func (m *ValueMap[K, V]) LoadOrStore(key K, value *V) (actual *V, loaded bool) {
	for {
		actual, loaded = m.Load(key)
		if loaded {
			return actual, true
		}
		mv := m.newValue(key, value)
		prev, loaded := m.m.LoadOrStore(key, mv)
		if !loaded {
			return value, false
		}
		mv.cleanup.Stop()
		_, ok := loadPointer(prev.value)
		if !ok {
			if m.m.CompareAndDelete(key, prev) {
				prev.cleanup.Stop()
			}
		}
	}
}

// CompareAndDelete is like [sync.Map.CompareAndDelete].
func (m *ValueMap[K, V]) CompareAndDelete(key K, old *V) (deleted bool) {
	for {
		mv, ok := m.m.Load(key)
		if !ok {
			return false
		}
		v, ok := loadPointer(mv.value)
		if !ok || v != old {
			return false
		}
		deleted = m.m.CompareAndDelete(key, mv)
		if deleted {
			mv.cleanup.Stop()
			return true
		}
	}
}

// CompareAndSwap is like [sync.Map.CompareAndSwap].
func (m *ValueMap[K, V]) CompareAndSwap(key K, oldValue, newValue *V) (swapped bool) {
	for {
		mv, ok := m.m.Load(key)
		if !ok {
			return false
		}
		v, ok := loadPointer(mv.value)
		if !ok || v != oldValue {
			return false
		}
		if oldValue == newValue {
			return true
		}
		newMv := m.newValue(key, newValue)
		swapped = m.m.CompareAndSwap(key, mv, newMv)
		if swapped {
			newMv = mv
		}
		newMv.cleanup.Stop()
		if swapped {
			return true
		}
	}
}

// Range is like [sync.Map.Range].
func (m *ValueMap[K, V]) Range(f func(key K, value *V) bool) {
	m.m.Range(func(k K, mv valueMapValue[V]) bool {
		v, ok := loadPointer(mv.value)
		return !ok || f(k, v)
	})
}

// All returns an iterator over all entries in the map.
// See [ValueMap.Range] for more details.
func (m *ValueMap[K, V]) All() iter.Seq2[K, *V] {
	return m.Range
}
