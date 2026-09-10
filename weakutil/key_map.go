package weakutil

import (
	"iter"
	"reflect"
	"runtime"
	"sync"
	"weak"

	"github.com/pierrre/go-libs/syncutil"
)

// KeyMap is a map that automatically evicts entries when the key is no longer reachable.
// It is safe for concurrent use.
// The zero value is ready to use.
// If a value is set with a nil key, it is never evicted.
//
// It implements the same methods as [sync.Map].
type KeyMap[K any, V any] struct {
	m                   syncutil.Map[weak.Pointer[K], keyMapEntry[V]]
	cleanupFunc         func(weak.Pointer[K])
	cleanupFuncOnce     sync.Once
	valueComparable     bool
	valueComparableOnce sync.Once
}

type keyMapEntry[V any] struct {
	value   V
	cleanup runtime.Cleanup
}

func (m *KeyMap[K, V]) getCleanupFunc() func(weak.Pointer[K]) {
	m.cleanupFuncOnce.Do(func() {
		m.cleanupFunc = m.cleanup
	})
	return m.cleanupFunc
}

func (m *KeyMap[K, V]) isValueComparable() bool {
	m.valueComparableOnce.Do(func() {
		m.valueComparable = isTypeSafelyComparable(reflect.TypeFor[V]())
	})
	return m.valueComparable
}

func (m *KeyMap[K, V]) newEntry(key *K, kp weak.Pointer[K], value V) (e keyMapEntry[V]) {
	e.value = value
	if key != nil {
		e.cleanup = runtime.AddCleanup(key, m.getCleanupFunc(), kp)
	}
	return e
}

func (m *KeyMap[K, V]) deleteEntry(kp weak.Pointer[K], e keyMapEntry[V]) (deleted bool) {
	e.cleanup.Stop()
	return m.m.CompareAndDelete(kp, e)
}

func (m *KeyMap[K, V]) deleteKey(kp weak.Pointer[K], e keyMapEntry[V]) {
	e.cleanup.Stop()
	m.m.Delete(kp)
}

func (m *KeyMap[K, V]) loadKey(kp weak.Pointer[K], e keyMapEntry[V]) (key *K, alive bool) {
	key, alive = loadPointer(kp)
	if !alive {
		m.deleteKey(kp, e)
	}
	return key, alive
}

func (m *KeyMap[K, V]) cleanup(kp weak.Pointer[K]) {
	m.m.Delete(kp)
}

// Store is like [sync.Map.Store].
func (m *KeyMap[K, V]) Store(key *K, value V) {
	_, _ = m.Swap(key, value)
}

// Load is like [sync.Map.Load].
func (m *KeyMap[K, V]) Load(key *K) (value V, ok bool) {
	kp := weak.Make(key)
	e, ok := m.m.Load(kp)
	return e.value, ok
}

// Delete is like [sync.Map.Delete].
func (m *KeyMap[K, V]) Delete(key *K) {
	_, _ = m.LoadAndDelete(key)
}

// Clear is like [sync.Map.Clear].
func (m *KeyMap[K, V]) Clear() {
	for {
		var count int64
		m.m.Range(func(kp weak.Pointer[K], e keyMapEntry[V]) bool {
			count++
			m.deleteKey(kp, e)
			return true
		})
		if count == 0 {
			return
		}
	}
}

// Swap is like [sync.Map.Swap].
func (m *KeyMap[K, V]) Swap(key *K, value V) (previous V, loaded bool) {
	kp := weak.Make(key)
	if m.isValueComparable() {
		e, ok := m.m.Load(kp)
		if ok && any(e.value) == any(value) {
			return e.value, true
		}
	}
	e := m.newEntry(key, kp, value)
	e, ok := m.m.Swap(kp, e)
	if ok {
		e.cleanup.Stop()
	}
	return e.value, ok
}

// LoadAndDelete is like [sync.Map.LoadAndDelete].
func (m *KeyMap[K, V]) LoadAndDelete(key *K) (value V, loaded bool) {
	kp := weak.Make(key)
	e, loaded := m.m.LoadAndDelete(kp)
	if loaded {
		e.cleanup.Stop()
	}
	return e.value, loaded
}

// LoadOrStore is like [sync.Map.LoadOrStore].
func (m *KeyMap[K, V]) LoadOrStore(key *K, value V) (actual V, loaded bool) {
	kp := weak.Make(key)
	var e keyMapEntry[V]
	for {
		e.cleanup.Stop()
		e, ok := m.m.Load(kp)
		if ok {
			return e.value, true
		}
		e = m.newEntry(key, kp, value)
		_, loaded = m.m.LoadOrStore(kp, e)
		if !loaded {
			return value, false
		}
	}
}

// CompareAndDelete is like [sync.Map.CompareAndDelete].
func (m *KeyMap[K, V]) CompareAndDelete(key *K, old V) (deleted bool) {
	kp := weak.Make(key)
	for {
		e, ok := m.m.Load(kp)
		if !ok {
			return false
		}
		if any(e.value) != any(old) {
			return false
		}
		if m.deleteEntry(kp, e) {
			return true
		}
	}
}

// CompareAndSwap is like [sync.Map.CompareAndSwap].
func (m *KeyMap[K, V]) CompareAndSwap(key *K, oldValue, newValue V) (swapped bool) {
	kp := weak.Make(key)
	for {
		e, ok := m.m.Load(kp)
		if !ok {
			return false
		}
		if any(e.value) != any(oldValue) {
			return false
		}
		if any(oldValue) == any(newValue) {
			return true
		}
		ne := m.newEntry(key, kp, newValue)
		swapped = m.m.CompareAndSwap(kp, e, ne)
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
func (m *KeyMap[K, V]) Range(f func(key *K, value V) bool) {
	m.m.Range(func(kp weak.Pointer[K], e keyMapEntry[V]) bool {
		key, alive := m.loadKey(kp, e)
		if !alive {
			return true
		}
		return f(key, e.value)
	})
}

// All returns an iterator over all entries in the map.
// See [KeyMap.Range] for more details.
func (m *KeyMap[K, V]) All() iter.Seq2[*K, V] {
	return m.Range
}
