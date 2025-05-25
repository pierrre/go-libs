package weakutil

import (
	"iter"
	"runtime"
	"slices"
	"sync"
	"sync/atomic"
	"weak"

	"github.com/pierrre/go-libs/syncutil"
)

// SimpleMap is a simple thread-safe map that holds [weak.Pointer] to values.
type SimpleMap[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]weak.Pointer[V]

	// OnGCDelete is called when a value is deleted from the map because it has been garbage collected.
	OnGCDelete func(key K)

	// AutoCleanInterval is the interval for auto-cleaning of the map (removing entries that have been garbage collected).
	// This interval is the number of calls to a method that modifies the map (such as [SimpleMap.Store], [SimpleMap.Delete], etc.), after which the map is cleaned.
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

	keysPool syncutil.ValuePool[[]K]
}

func (m *SimpleMap[K, V]) ensureInit() {
	if m.m == nil {
		m.m = make(map[K]weak.Pointer[V])
	}
}

func (m *SimpleMap[K, V]) len() int {
	m.mu.RLock()
	l := len(m.m)
	m.mu.RUnlock()
	return l
}

func (m *SimpleMap[K, V]) get(key K) (pointer weak.Pointer[V], ok bool) {
	m.mu.RLock()
	pointer, ok = m.m[key]
	m.mu.RUnlock()
	return pointer, ok
}

func (m *SimpleMap[K, V]) resolve(pointer weak.Pointer[V]) (value *V, ok bool) {
	if pointer == (weak.Pointer[V]{}) {
		// The value was set to nil.
		return nil, true
	}
	value = pointer.Value()
	return value, value != nil
}

func (m *SimpleMap[K, V]) resolveClean(key K, pointer weak.Pointer[V]) (value *V, ok bool) {
	value, ok = m.resolve(pointer)
	if ok {
		return value, true
	}
	// The value has been garbage collected, delete it from the map.
	m.compareAndDelete(key, pointer)
	return nil, false
}

func (m *SimpleMap[K, V]) getResolveClean(key K) (value *V, ok bool) {
	pointer, ok := m.get(key)
	if !ok {
		return nil, false
	}
	return m.resolveClean(key, pointer)
}

func (m *SimpleMap[K, V]) compareAndDelete(key K, oldPointer weak.Pointer[V]) (deleted bool) {
	deleted = false
	m.mu.Lock()
	currentPointer, ok := m.m[key]
	if ok && currentPointer == oldPointer {
		delete(m.m, key)
		deleted = true
	}
	m.mu.Unlock()
	return deleted
}

// Store is like [sync.Map.Store].
func (m *SimpleMap[K, V]) Store(key K, value *V) {
	m.autoClean()
	newPointer := weak.Make(value)
	oldPointer, ok := m.get(key)
	if ok && newPointer == oldPointer {
		return
	}
	m.mu.Lock()
	m.ensureInit()
	m.m[key] = newPointer
	m.mu.Unlock()
}

// Load is like [sync.Map.Load].
func (m *SimpleMap[K, V]) Load(key K) (value *V, ok bool) {
	return m.getResolveClean(key)
}

// Delete is like [sync.Map.Delete].
func (m *SimpleMap[K, V]) Delete(key K) {
	m.autoClean()
	// Fast path
	_, ok := m.get(key)
	if !ok {
		return
	}
	// Slow path
	m.mu.Lock()
	m.ensureInit()
	delete(m.m, key)
	m.mu.Unlock()
}

// Clear is like [sync.Map.Clear].
func (m *SimpleMap[K, V]) Clear() {
	// Fast path
	if m.len() == 0 {
		return
	}
	// Slow path
	m.mu.Lock()
	m.ensureInit()
	clear(m.m)
	m.mu.Unlock()
}

// LoadAndDelete is like [sync.Map.LoadAndDelete].
func (m *SimpleMap[K, V]) LoadAndDelete(key K) (value *V, loaded bool) {
	m.autoClean()
	// Fast path
	_, loaded = m.getResolveClean(key)
	if !loaded {
		return nil, false
	}
	// Slow path
	m.mu.Lock()
	pointer, loaded := m.m[key]
	if loaded {
		delete(m.m, key)
	}
	m.mu.Unlock()
	if !loaded {
		return nil, false
	}
	return m.resolveClean(key, pointer)
}

// LoadOrStore is like [sync.Map.LoadOrStore].
func (m *SimpleMap[K, V]) LoadOrStore(key K, value *V) (actual *V, loaded bool) {
	m.autoClean()
	// Fast path
	actual, loaded = m.getResolveClean(key)
	if loaded {
		return actual, true
	}
	// Slow path
	newPointer := weak.Make(value)
	m.mu.Lock()
	oldPointer, loaded := m.m[key]
	if loaded {
		actual, loaded = m.resolve(oldPointer)
		if !loaded {
			m.ensureInit()
			m.m[key] = newPointer
			actual = value
		}
	}
	m.mu.Unlock()
	return actual, false
}

// Swap is like [sync.Map.Swap].
func (m *SimpleMap[K, V]) Swap(key K, value *V) (previous *V, loaded bool) {
	m.autoClean()
	// Always slow path
	newPointer := weak.Make(value)
	m.mu.Lock()
	m.ensureInit()
	oldPointer, loaded := m.m[key]
	m.m[key] = newPointer
	m.mu.Unlock()
	if !loaded {
		return nil, false
	}
	return m.resolveClean(key, oldPointer)
}

// CompareAndDelete is like [sync.Map.CompareAndDelete].
func (m *SimpleMap[K, V]) CompareAndDelete(key K, old *V) (deleted bool) {
	m.autoClean()
	oldPointer := weak.Make(old)
	// Fast path
	pointer, ok := m.get(key)
	if !ok || pointer != oldPointer {
		return false
	}
	// Slow path
	return m.compareAndDelete(key, oldPointer)
}

// CompareAndSwap is like [sync.Map.CompareAndSwap].
func (m *SimpleMap[K, V]) CompareAndSwap(key K, oldValue, newValue *V) (swapped bool) {
	m.autoClean()
	oldPointer := weak.Make(oldValue)
	// Fast path
	pointer, ok := m.get(key)
	if !ok || pointer != oldPointer {
		return false
	}
	// Slow path
	newPointer := weak.Make(newValue)
	swapped = false
	m.mu.Lock()
	if m.m[key] == oldPointer {
		m.m[key] = newPointer
		swapped = true
	}
	m.mu.Unlock()
	return swapped
}

// Range is like [sync.Map.Range].
func (m *SimpleMap[K, V]) Range(f func(key K, value *V) bool) {
	// Fast path
	if m.len() == 0 {
		return
	}
	// Slow path
	keys := m.keysPool.Get()
	defer func() {
		m.keysPool.Put(keys)
	}()
	keys = keys[:0]
	m.mu.RLock()
	keys = slices.Grow(keys, len(m.m))
	for key := range m.m {
		keys = append(keys, key)
	}
	m.mu.RUnlock()
	for _, key := range keys {
		pointer, ok := m.get(key)
		if !ok {
			continue
		}
		value, ok := m.resolveClean(key, pointer)
		if !ok {
			continue
		}
		if !f(key, value) {
			return
		}
	}
}

// All returns an iterator over all entries in the map.
func (m *SimpleMap[K, V]) All() iter.Seq2[K, *V] {
	return m.Range
}

// Clean removes all entries from the map that have been garbage collected.
func (m *SimpleMap[K, V]) Clean() {
	m.clean()
}

func (m *SimpleMap[K, V]) clean() (count int64) {
	for k, v := range m.Range {
		runtime.KeepAlive(k)
		runtime.KeepAlive(v)
		count++
	}
	return count
}

func (m *SimpleMap[K, V]) autoClean() {
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
