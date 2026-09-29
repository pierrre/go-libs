package weakutil

import (
	"iter"
	"reflect"
	"runtime"
	"weak"

	"github.com/pierrre/go-libs/reflectutil"
	"github.com/pierrre/go-libs/syncutil"
)

// KeyMap is a map that automatically evicts entries when the key is no longer reachable.
// It is safe for concurrent use.
// The zero value is ready to use.
// If a value is set with a nil key, it is never evicted.
// Additionally, when cleanup is disabled, dead entries are removed from the map every [KeyMap.GetSweepWriteCount] writes.
//
// It implements the same methods as [sync.Map].
type KeyMap[K any, V any] struct {
	commonMap[weak.Pointer[K], keyMapEntry[V]]
	cleanupFunc     lazyValue[func(keyMapCleanupArg[K, V])]
	valueComparable lazyValue[bool]
}

type keyMapEntry[V any] struct {
	value   V
	cleanup runtime.Cleanup
}

func (e keyMapEntry[V]) stopCleanup() {
	e.cleanup.Stop()
}

func (m *KeyMap[K, V]) getCleanupFunc() func(keyMapCleanupArg[K, V]) {
	return m.cleanupFunc.get(func() func(keyMapCleanupArg[K, V]) {
		return keyMapCleanup
	})
}

func (m *KeyMap[K, V]) isValueComparable() bool {
	return m.valueComparable.get(func() bool {
		return reflectutil.IsTypeStrictlyComparable(reflect.TypeFor[V]())
	})
}

func (m *KeyMap[K, V]) newEntry(key *K, kp weak.Pointer[K], value V) (e keyMapEntry[V]) {
	e.value = value
	if key != nil {
		if m.IsCleanupEnabled() {
			e.cleanup = runtime.AddCleanup(key, m.getCleanupFunc(), keyMapCleanupArg[K, V]{
				m:   m.m,
				key: kp,
			})
		}
	}
	return e
}

func (m *KeyMap[K, V]) deleteEntry(kp weak.Pointer[K], e keyMapEntry[V]) (deleted bool) {
	e.cleanup.Stop()
	if m.isValueComparable() {
		return m.m.CompareAndDelete(kp, e)
	}
	m.m.Delete(kp)
	return true
}

func (m *KeyMap[K, V]) loadKey(kp weak.Pointer[K], e keyMapEntry[V]) (*K, bool) {
	return loadWeak(m, kp, e, kp)
}

type keyMapCleanupArg[K any, V any] struct {
	m   *syncutil.Map[weak.Pointer[K], keyMapEntry[V]]
	key weak.Pointer[K]
}

func keyMapCleanup[K any, V any](arg keyMapCleanupArg[K, V]) {
	arg.m.Delete(arg.key)
}

func (m *KeyMap[K, V]) sweep() {
	m.Range(func(_ *K, _ V) bool {
		return true
	})
}

// Store is like [sync.Map.Store].
func (m *KeyMap[K, V]) Store(key *K, value V) {
	_, _ = m.Swap(key, value)
}

// Load is like [sync.Map.Load].
func (m *KeyMap[K, V]) Load(key *K) (value V, ok bool) {
	m.ensureInit()
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
	m.ensureInit()
	clearMap(m.m, m.deleteEntry)
}

// Swap is like [sync.Map.Swap].
func (m *KeyMap[K, V]) Swap(key *K, value V) (previous V, loaded bool) {
	m.ensureInit()
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
	m.maybeSweep(m.sweep)
	return e.value, ok
}

// LoadAndDelete is like [sync.Map.LoadAndDelete].
func (m *KeyMap[K, V]) LoadAndDelete(key *K) (value V, loaded bool) {
	m.ensureInit()
	kp := weak.Make(key)
	e, loaded := m.m.LoadAndDelete(kp)
	if loaded {
		e.cleanup.Stop()
		m.maybeSweep(m.sweep)
	}
	return e.value, loaded
}

// LoadOrStore is like [sync.Map.LoadOrStore].
func (m *KeyMap[K, V]) LoadOrStore(key *K, value V) (actual V, loaded bool) {
	m.ensureInit()
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
			m.maybeSweep(m.sweep)
			return value, false
		}
	}
}

// CompareAndDelete is like [sync.Map.CompareAndDelete].
func (m *KeyMap[K, V]) CompareAndDelete(key *K, old V) (deleted bool) {
	m.ensureInit()
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
			m.maybeSweep(m.sweep)
			return true
		}
	}
}

// CompareAndSwap is like [sync.Map.CompareAndSwap].
func (m *KeyMap[K, V]) CompareAndSwap(key *K, oldValue, newValue V) (swapped bool) {
	m.ensureInit()
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
			m.maybeSweep(m.sweep)
			return true
		}
	}
}

// Range is like [sync.Map.Range].
func (m *KeyMap[K, V]) Range(f func(key *K, value V) bool) {
	m.ensureInit()
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
