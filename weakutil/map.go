package weakutil

import (
	"iter"
	"runtime"
	"sync"
	"sync/atomic"
	"weak"

	"github.com/pierrre/go-libs/syncutil"
)

// Map is a [syncutil.Map] that holds weak pointers to values.
// Values are automatically removed from the map when they are garbage collected.
//
// The map can only store pointers to values.
// The map can store nil pointers, and they're never garbage collected.
// Methods such as [Map.Load] may return nil pointers.
//
// The zero value of Map is ready to use.
type Map[K any, V any] struct {
	m syncutil.Map[K, weak.Pointer[V]]

	// OnGCDelete is called when a value is deleted from the map because it has been garbage collected.
	OnGCDelete func(key K)

	// AutoCleanInterval is the interval for auto-cleaning of the map (removing entries that have been garbage collected).
	// This interval is the number of calls to a method that modifies the map (such as [Map.Store], [Map.Delete], etc.), after which the map is cleaned.
	// The 0 value is equivalent to the number of entries in the map after the previous auto-cleaning (or 1), so it scales automatically with the number of entries in the map.
	// If it is less than 0, auto-cleaning is disabled.
	// Auto-cleaning is skipped if it's already running.
	AutoCleanInterval int64
	// AutoCleanAsync indicates whether auto-cleaning should be done asynchronously (in a goroutine), instead of blocking the caller.
	AutoCleanAsync bool
	// OnAutoClean is called when auto-cleaning is performed.
	OnAutoClean        func()
	autoCleanCounter   int64
	autoCleanLock      sync.Mutex
	autoCleanLastCount int64
}

func (m *Map[K, V]) resolveClean(key K, pointer weak.Pointer[V]) (value *V, ok bool) {
	if pointer == (weak.Pointer[V]{}) {
		// The value was set to nil.
		return nil, true
	}
	value = pointer.Value()
	if value != nil {
		// The value is still alive.
		return value, true
	}
	// The value has been garbage collected, delete it from the map.
	m.m.CompareAndDelete(key, pointer)
	if m.OnGCDelete != nil {
		m.OnGCDelete(key)
	}
	return nil, false
}

// Store is a wrapper around [sync.Map.Store].
func (m *Map[K, V]) Store(key K, value *V) {
	m.autoClean()
	newPointer := weak.Make(value)
	oldPointer, ok := m.m.Load(key)
	if ok && newPointer == oldPointer {
		return
	}
	m.m.Store(key, newPointer)
}

// Load is a wrapper around [sync.Map.Load].
func (m *Map[K, V]) Load(key K) (value *V, ok bool) {
	pointer, ok := m.m.Load(key)
	if !ok {
		return nil, false
	}
	return m.resolveClean(key, pointer)
}

// Delete is a wrapper around [sync.Map.Delete].
func (m *Map[K, V]) Delete(key K) {
	m.autoClean()
	m.m.Delete(key)
}

// Clear is a wrapper around [sync.Map.Clear].
func (m *Map[K, V]) Clear() {
	m.m.Clear()
}

// LoadAndDelete is a wrapper around [sync.Map.LoadAndDelete].
func (m *Map[K, V]) LoadAndDelete(key K) (value *V, loaded bool) {
	m.autoClean()
	pointer, loaded := m.m.LoadAndDelete(key)
	if !loaded {
		return nil, false
	}
	return m.resolveClean(key, pointer)
}

// LoadOrStore is a wrapper around [sync.Map.LoadOrStore].
func (m *Map[K, V]) LoadOrStore(key K, value *V) (actual *V, loaded bool) {
	m.autoClean()
	pointer := weak.Make(value)
	for {
		actualPointer, loaded := m.m.LoadOrStore(key, pointer)
		actual, ok := m.resolveClean(key, actualPointer)
		if ok {
			return actual, loaded
		}
	}
}

// Swap is a wrapper around [sync.Map.Swap].
func (m *Map[K, V]) Swap(key K, value *V) (previous *V, loaded bool) {
	m.autoClean()
	pointer := weak.Make(value)
	pointer, loaded = m.m.Swap(key, pointer)
	if !loaded {
		return nil, false
	}
	return m.resolveClean(key, pointer)
}

// CompareAndDelete is a wrapper around [sync.Map.CompareAndDelete].
func (m *Map[K, V]) CompareAndDelete(key K, old *V) (deleted bool) {
	m.autoClean()
	pointer := weak.Make(old)
	return m.m.CompareAndDelete(key, pointer)
}

// CompareAndSwap is a wrapper around [sync.Map.CompareAndSwap].
func (m *Map[K, V]) CompareAndSwap(key K, oldValue, newValue *V) (swapped bool) {
	m.autoClean()
	oldPointer := weak.Make(oldValue)
	newPointer := weak.Make(newValue)
	return m.m.CompareAndSwap(key, oldPointer, newPointer)
}

// Range is a wrapper around [sync.Map.Range].
func (m *Map[K, V]) Range(f func(key K, value *V) bool) {
	m.m.Range(func(key K, pointer weak.Pointer[V]) bool {
		value, ok := m.resolveClean(key, pointer)
		if !ok {
			return true
		}
		return f(key, value)
	})
}

// All returns an iterator over all entries in the map.
func (m *Map[K, V]) All() iter.Seq2[K, *V] {
	return m.Range
}

// Clean removes all entries from the map that have been garbage collected.
func (m *Map[K, V]) Clean() {
	m.clean()
}

func (m *Map[K, V]) clean() (count int64) {
	for k, v := range m.Range {
		runtime.KeepAlive(k)
		runtime.KeepAlive(v)
		count++
	}
	return count
}

func (m *Map[K, V]) autoClean() {
	interval := m.AutoCleanInterval
	if interval < 0 {
		return
	}
	if interval == 0 {
		interval = m.autoCleanLastCount
	}
	if interval <= 0 {
		interval = 1
	}
	counter := atomic.AddInt64(&m.autoCleanCounter, 1)
	if counter%interval != 0 {
		return
	}
	ok := m.autoCleanLock.TryLock()
	if !ok {
		return
	}
	f := func() {
		defer m.autoCleanLock.Unlock()
		if m.OnAutoClean != nil {
			m.OnAutoClean()
		}
		m.autoCleanLastCount = m.clean()
	}
	if m.AutoCleanAsync {
		go f()
	} else {
		f()
	}
}
