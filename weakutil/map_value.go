package weakutil

import (
	"iter"
	"runtime"
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
	m           syncutil.Map[K, valueMapEntry[V]]
	cleanupFunc lazyValue[func(valueMapCleanupArg[K, V])]
}

type valueMapEntry[T any] struct {
	value   weak.Pointer[T]
	cleanup runtime.Cleanup
}

func (m *ValueMap[K, V]) getCleanupFunc() func(valueMapCleanupArg[K, V]) {
	return m.cleanupFunc.get(func() func(valueMapCleanupArg[K, V]) {
		return m.cleanup
	})
}

func (m *ValueMap[K, V]) newEntry(key K, value *V) valueMapEntry[V] {
	var e valueMapEntry[V]
	if value != nil {
		e.value = weak.Make(value)
		e.cleanup = runtime.AddCleanup(value, m.getCleanupFunc(), valueMapCleanupArg[K, V]{
			key:   key,
			value: e.value,
		})
	}
	return e
}

func (m *ValueMap[K, V]) deleteEntry(key K, e valueMapEntry[V]) (deleted bool) {
	e.cleanup.Stop()
	return m.m.CompareAndDelete(key, e)
}

func (m *ValueMap[K, V]) loadValue(key K, e valueMapEntry[V]) (value *V, alive bool) {
	value, alive = loadPointer(e.value)
	if !alive {
		m.deleteEntry(key, e)
	}
	return value, alive
}

type valueMapCleanupArg[K comparable, V any] struct {
	key   K
	value weak.Pointer[V]
}

func (m *ValueMap[K, V]) cleanup(mc valueMapCleanupArg[K, V]) {
	e, ok := m.m.Load(mc.key)
	if ok && e.value == mc.value {
		m.m.CompareAndDelete(mc.key, e)
	}
}

// Store is like [sync.Map.Store].
func (m *ValueMap[K, V]) Store(key K, value *V) {
	_, _ = m.Swap(key, value)
}

// Load is like [sync.Map.Load].
func (m *ValueMap[K, V]) Load(key K) (value *V, ok bool) {
	e, ok := m.m.Load(key)
	if !ok {
		return nil, false
	}
	value, ok = m.loadValue(key, e)
	return value, ok
}

// Delete is like [sync.Map.Delete].
func (m *ValueMap[K, V]) Delete(key K) {
	_, _ = m.LoadAndDelete(key)
}

// Clear is like [sync.Map.Clear].
func (m *ValueMap[K, V]) Clear() {
	clearMap(&m.m, m.deleteEntry)
}

// Swap is like [sync.Map.Swap].
func (m *ValueMap[K, V]) Swap(key K, value *V) (previous *V, loaded bool) {
	previous, loaded = m.Load(key)
	if loaded && previous == value {
		return previous, true
	}
	e := m.newEntry(key, value)
	e, ok := m.m.Swap(key, e)
	if ok {
		previous, loaded = loadPointer(e.value)
		e.cleanup.Stop()
	}
	return previous, loaded
}

// LoadAndDelete is like [sync.Map.LoadAndDelete].
func (m *ValueMap[K, V]) LoadAndDelete(key K) (value *V, loaded bool) {
	e, ok := m.m.LoadAndDelete(key)
	if ok {
		value, loaded = loadPointer(e.value)
		e.cleanup.Stop()
	}
	return value, loaded
}

// LoadOrStore is like [sync.Map.LoadOrStore].
func (m *ValueMap[K, V]) LoadOrStore(key K, value *V) (actual *V, loaded bool) {
	for {
		e, ok := m.m.Load(key)
		if ok {
			actual, loaded = loadPointer(e.value)
			if loaded {
				return actual, true
			}
		}
		ne := m.newEntry(key, value)
		prev, loaded := m.m.LoadOrStore(key, ne)
		if !loaded {
			return value, false
		}
		ne.cleanup.Stop()
		m.loadValue(key, prev)
	}
}

// CompareAndDelete is like [sync.Map.CompareAndDelete].
func (m *ValueMap[K, V]) CompareAndDelete(key K, old *V) (deleted bool) {
	for {
		e, ok := m.m.Load(key)
		if !ok {
			return false
		}
		v, alive := m.loadValue(key, e)
		if !alive {
			return false
		}
		if v != old {
			return false
		}
		if m.deleteEntry(key, e) {
			return true
		}
	}
}

// CompareAndSwap is like [sync.Map.CompareAndSwap].
func (m *ValueMap[K, V]) CompareAndSwap(key K, oldValue, newValue *V) (swapped bool) {
	for {
		e, ok := m.m.Load(key)
		if !ok {
			return false
		}
		v, alive := m.loadValue(key, e)
		if !alive {
			return false
		}
		if v != oldValue {
			return false
		}
		if oldValue == newValue {
			return true
		}
		ne := m.newEntry(key, newValue)
		swapped = m.m.CompareAndSwap(key, e, ne)
		if swapped {
			ne = e
		}
		ne.cleanup.Stop()
		if swapped {
			return true
		}
	}
}

// Range is like [sync.Map.Range].
func (m *ValueMap[K, V]) Range(f func(key K, value *V) bool) {
	m.m.Range(func(key K, e valueMapEntry[V]) bool {
		value, alive := m.loadValue(key, e)
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
