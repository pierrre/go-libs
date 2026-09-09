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
type ValueMap[K comparable, V any] struct {
	m               syncutil.Map[K, mapValue[V]]
	cleanupFunc     func(mapCleanup[K, V])
	cleanupFuncOnce sync.Once
}

type mapValue[T any] struct {
	pointer weak.Pointer[T]
	cleanup runtime.Cleanup
}

func (mv mapValue[T]) get() (*T, bool) {
	return loadPointer(mv.pointer)
}

func (mv mapValue[T]) stopCleanup() {
	if mv.pointer != (weak.Pointer[T]{}) {
		mv.cleanup.Stop()
	}
}

func (m *ValueMap[K, V]) getCleanupFunc() func(mapCleanup[K, V]) {
	m.cleanupFuncOnce.Do(func() {
		m.cleanupFunc = m.cleanup
	})
	return m.cleanupFunc
}

func (m *ValueMap[K, V]) newValue(key K, value *V) mapValue[V] {
	var mv mapValue[V]
	if value != nil {
		mv.pointer = weak.Make(value)
		mv.cleanup = runtime.AddCleanup(value, m.getCleanupFunc(), mapCleanup[K, V]{
			key:     key,
			pointer: mv.pointer,
		})
	}
	return mv
}

type mapCleanup[K comparable, V any] struct {
	key     K
	pointer weak.Pointer[V]
}

func (m *ValueMap[K, V]) cleanup(mc mapCleanup[K, V]) {
	mv, ok := m.m.Load(mc.key)
	if ok && mv.pointer == mc.pointer {
		m.m.CompareAndDelete(mc.key, mv)
	}
}

// Store is like [sync.Map.Store].
func (m *ValueMap[K, V]) Store(key K, value *V) {
	v, ok := m.Load(key)
	if ok && v == value {
		return
	}
	mv := m.newValue(key, value)
	mv, ok = m.m.Swap(key, mv)
	if ok {
		mv.stopCleanup()
	}
}

// Load is like [sync.Map.Load].
func (m *ValueMap[K, V]) Load(key K) (value *V, ok bool) {
	mv, ok := m.m.Load(key)
	if !ok {
		return nil, false
	}
	return mv.get()
}

// Delete is like [sync.Map.Delete].
func (m *ValueMap[K, V]) Delete(key K) {
	mv, ok := m.m.LoadAndDelete(key)
	if ok {
		mv.stopCleanup()
	}
}

// Clear is like [sync.Map.Clear].
func (m *ValueMap[K, V]) Clear() {
	m.m.Range(func(k K, mv mapValue[V]) bool {
		mv.stopCleanup()
		return true
	})
	m.m.Clear()
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
		previous, loaded = mv.get()
		mv.stopCleanup()
	}
	return previous, loaded
}

// LoadAndDelete is like [sync.Map.LoadAndDelete].
func (m *ValueMap[K, V]) LoadAndDelete(key K) (value *V, loaded bool) {
	mv, ok := m.m.LoadAndDelete(key)
	if ok {
		value, loaded = mv.get()
		mv.stopCleanup()
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
		mv.stopCleanup()
		_, ok := prev.get()
		if !ok {
			if m.m.CompareAndDelete(key, prev) {
				prev.stopCleanup()
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
		v, ok := mv.get()
		if !ok || v != old {
			return false
		}
		deleted = m.m.CompareAndDelete(key, mv)
		if deleted {
			mv.stopCleanup()
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
		v, ok := mv.get()
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
		newMv.stopCleanup()
		if swapped {
			return true
		}
	}
}

// Range is like [sync.Map.Range].
func (m *ValueMap[K, V]) Range(f func(key K, value *V) bool) {
	m.m.Range(func(k K, mv mapValue[V]) bool {
		v, ok := mv.get()
		return !ok || f(k, v)
	})
}

// All returns an iterator over all entries in the map.
// See [ValueMap.Range] for more details.
func (m *ValueMap[K, V]) All() iter.Seq2[K, *V] {
	return m.Range
}
