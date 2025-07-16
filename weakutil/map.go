package weakutil

import (
	"iter"
	"runtime"
	"weak"

	"github.com/pierrre/go-libs/syncutil"
)

// Map is a map that stores [weak.Pointer].
// Values are automatically evicted when they are no longer reachable.
// It is safe for concurrent use.
// The zero value is ready to use.
//
// It implements the same methods as [sync.Map].
type Map[K comparable, V any] struct {
	m syncutil.Map[K, mapValue[V]]
}

type mapValue[T any] struct {
	pointer weak.Pointer[T]
	cleanup runtime.Cleanup
}

func (mv mapValue[T]) get() (*T, bool) {
	if mv.pointer == (weak.Pointer[T]{}) {
		return nil, true
	}
	v := mv.pointer.Value()
	return v, v != nil
}

func (mv mapValue[T]) stopCleanup() {
	if mv.pointer != (weak.Pointer[T]{}) {
		mv.cleanup.Stop()
	}
}

func (m *Map[K, V]) newValue(key K, value *V) mapValue[V] {
	var mv mapValue[V]
	if value != nil {
		mv.pointer = weak.Make(value)
		mv.cleanup = runtime.AddCleanup(value, m.cleanup, mapCleanup[K, V]{
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

func (m *Map[K, V]) cleanup(mc mapCleanup[K, V]) {
	mv, ok := m.m.Load(mc.key)
	if ok && mv.pointer == mc.pointer {
		m.m.CompareAndDelete(mc.key, mv)
	}
}

// Store is like [sync.Map.Store].
func (m *Map[K, V]) Store(key K, value *V) {
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
func (m *Map[K, V]) Load(key K) (value *V, ok bool) {
	mv, ok := m.m.Load(key)
	if !ok {
		return nil, false
	}
	return mv.get()
}

// Delete is like [sync.Map.Delete].
func (m *Map[K, V]) Delete(key K) {
	mv, ok := m.m.LoadAndDelete(key)
	if ok {
		mv.stopCleanup()
	}
}

// Clear is like [sync.Map.Clear].
func (m *Map[K, V]) Clear() {
	for k, mv := range m.m.Range {
		m.m.CompareAndDelete(k, mv)
		mv.stopCleanup()
	}
}

// Swap is like [sync.Map.Swap].
func (m *Map[K, V]) Swap(key K, value *V) (previous *V, loaded bool) {
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
func (m *Map[K, V]) LoadAndDelete(key K) (value *V, loaded bool) {
	mv, ok := m.m.LoadAndDelete(key)
	if ok {
		value, loaded = mv.get()
		mv.stopCleanup()
	}
	return value, loaded
}

// LoadOrStore is like [sync.Map.LoadOrStore].
func (m *Map[K, V]) LoadOrStore(key K, value *V) (actual *V, loaded bool) {
	var mv mapValue[V]
	for {
		mv.stopCleanup()
		actual, loaded = m.Load(key)
		if loaded {
			return actual, true
		}
		mv = m.newValue(key, value)
		_, loaded = m.m.LoadOrStore(key, mv)
		if !loaded {
			return value, false
		}
	}
}

// CompareAndDelete is like [sync.Map.CompareAndDelete].
func (m *Map[K, V]) CompareAndDelete(key K, old *V) (deleted bool) {
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
func (m *Map[K, V]) CompareAndSwap(key K, oldValue, newValue *V) (swapped bool) {
	for {
		mv, ok := m.m.Load(key)
		if !ok {
			return false
		}
		v, ok := mv.get()
		if !ok || v != oldValue {
			return false
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
func (m *Map[K, V]) Range(f func(key K, value *V) bool) {
	for k, mv := range m.m.Range {
		v, ok := mv.get()
		if ok && !f(k, v) {
			return
		}
	}
}

// All returns an iterator over all entries in the map.
// See [Map.Range] for more details.
func (m *Map[K, V]) All() iter.Seq2[K, *V] {
	return m.Range
}
