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
			key:   key,
			value: mv.value,
		})
	}
	return mv
}

func (m *ValueMap[K, V]) deleteValue(key K, mv valueMapValue[V]) (deleted bool) {
	mv.cleanup.Stop()
	return m.m.CompareAndDelete(key, mv)
}

func (m *ValueMap[K, V]) loadValue(key K, mv valueMapValue[V]) (value *V, alive bool) {
	value, alive = loadPointer(mv.value)
	if !alive {
		m.deleteValue(key, mv)
	}
	return value, alive
}

type valueMapCleanupArg[K comparable, V any] struct {
	key   K
	value weak.Pointer[V]
}

func (m *ValueMap[K, V]) cleanup(mc valueMapCleanupArg[K, V]) {
	mv, ok := m.m.Load(mc.key)
	if ok && mv.value == mc.value {
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
	value, ok = m.loadValue(key, mv)
	return value, ok
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
			m.deleteValue(k, mv)
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
		mv, ok := m.m.Load(key)
		if ok {
			actual, loaded = loadPointer(mv.value)
			if loaded {
				return actual, true
			}
		}
		newMv := m.newValue(key, value)
		prev, loaded := m.m.LoadOrStore(key, newMv)
		if !loaded {
			return value, false
		}
		newMv.cleanup.Stop()
		m.loadValue(key, prev)
	}
}

// CompareAndDelete is like [sync.Map.CompareAndDelete].
func (m *ValueMap[K, V]) CompareAndDelete(key K, old *V) (deleted bool) {
	for {
		mv, ok := m.m.Load(key)
		if !ok {
			return false
		}
		v, alive := m.loadValue(key, mv)
		if !alive {
			return false
		}
		if v != old {
			return false
		}
		if m.deleteValue(key, mv) {
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
		v, alive := m.loadValue(key, mv)
		if !alive {
			return false
		}
		if v != oldValue {
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
	m.m.Range(func(key K, mv valueMapValue[V]) bool {
		value, alive := m.loadValue(key, mv)
		if !alive {
			return true
		}
		return f(key, value)
	})
}

// All returns an iterator over all entries in the map.
// See [ValueMap.Range] for more details.
func (m *ValueMap[K, V]) All() iter.Seq2[K, *V] {
	return m.Range
}
