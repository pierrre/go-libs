package weakutil

import (
	"iter"
	"runtime"
	"weak"

	"github.com/pierrre/go-libs/syncutil"
)

// KeyValueMap is a map that automatically evicts entries when the key or the value is no longer reachable.
// It is safe for concurrent use.
// The zero value is ready to use.
// A nil key or a nil value never triggers eviction by itself.
// With a nil value, the entry is still evicted when the key is no longer reachable.
// With a nil key, the entry is still evicted when the value is no longer reachable.
// If both are nil, the entry is never evicted.
// Additionally, when cleanup is disabled, dead entries are removed from the map every [KeyValueMap.GetSweepWriteCount] writes.
//
// It implements the same methods as [sync.Map].
type KeyValueMap[K any, V any] struct {
	commonMap[weak.Pointer[K], keyValueMapEntry[K, V]]
	keyCleanupFunc   lazyValue[func(keyValueMapKeyCleanupArg[K, V])]
	valueCleanupFunc lazyValue[func(keyValueMapCleanupArg[K, V])]
}

type keyValueMapEntry[K any, V any] struct {
	value        weak.Pointer[V]
	keyCleanup   runtime.Cleanup
	valueCleanup runtime.Cleanup
}

func (e keyValueMapEntry[K, V]) stopCleanup() {
	e.keyCleanup.Stop()
	e.valueCleanup.Stop()
}

type keyValueMapKeyCleanupArg[K any, V any] struct {
	m   *syncutil.Map[weak.Pointer[K], keyValueMapEntry[K, V]]
	key weak.Pointer[K]
}

type keyValueMapCleanupArg[K any, V any] struct {
	m     *syncutil.Map[weak.Pointer[K], keyValueMapEntry[K, V]]
	key   weak.Pointer[K]
	value weak.Pointer[V]
}

func (m *KeyValueMap[K, V]) getKeyCleanupFunc() func(keyValueMapKeyCleanupArg[K, V]) {
	return m.keyCleanupFunc.get(func() func(keyValueMapKeyCleanupArg[K, V]) {
		return keyValueMapKeyCleanup
	})
}

func (m *KeyValueMap[K, V]) getValueCleanupFunc() func(keyValueMapCleanupArg[K, V]) {
	return m.valueCleanupFunc.get(func() func(keyValueMapCleanupArg[K, V]) {
		return keyValueMapValueCleanup
	})
}

func (m *KeyValueMap[K, V]) newEntry(key *K, kp weak.Pointer[K], value *V) (e keyValueMapEntry[K, V]) {
	cleanupEnabled := m.IsCleanupEnabled()
	if key != nil {
		if cleanupEnabled {
			e.keyCleanup = runtime.AddCleanup(key, m.getKeyCleanupFunc(), keyValueMapKeyCleanupArg[K, V]{
				m:   m.m,
				key: kp,
			})
		}
	}
	if value != nil {
		vp := weak.Make(value)
		e.value = vp
		if cleanupEnabled {
			e.valueCleanup = runtime.AddCleanup(value, m.getValueCleanupFunc(), keyValueMapCleanupArg[K, V]{
				m:     m.m,
				key:   kp,
				value: vp,
			})
		}
	}
	return e
}

func (m *KeyValueMap[K, V]) deleteEntry(kp weak.Pointer[K], e keyValueMapEntry[K, V]) (deleted bool) {
	e.stopCleanup()
	return m.m.CompareAndDelete(kp, e)
}

func (m *KeyValueMap[K, V]) loadValue(kp weak.Pointer[K], e keyValueMapEntry[K, V]) (*V, bool) {
	return loadWeak(m, kp, e, e.value)
}

func (m *KeyValueMap[K, V]) loadKey(kp weak.Pointer[K], e keyValueMapEntry[K, V]) (*K, bool) {
	return loadWeak(m, kp, e, kp)
}

func keyValueMapKeyCleanup[K any, V any](arg keyValueMapKeyCleanupArg[K, V]) {
	e, ok := arg.m.LoadAndDelete(arg.key)
	if ok {
		_, alive := loadPointer(e.value)
		if alive {
			e.valueCleanup.Stop()
		}
	}
}

func keyValueMapValueCleanup[K any, V any](arg keyValueMapCleanupArg[K, V]) {
	e, ok := arg.m.Load(arg.key)
	if ok && e.value == arg.value {
		arg.m.CompareAndDelete(arg.key, e)
		_, alive := loadPointer(arg.key)
		if alive {
			e.keyCleanup.Stop()
		}
	}
}

func (m *KeyValueMap[K, V]) sweep() {
	m.Range(func(_ *K, _ *V) bool {
		return true
	})
}

// Store is like [sync.Map.Store].
func (m *KeyValueMap[K, V]) Store(key *K, value *V) {
	_, _ = m.Swap(key, value)
}

// Load is like [sync.Map.Load].
func (m *KeyValueMap[K, V]) Load(key *K) (value *V, ok bool) {
	m.ensureInit()
	kp := weak.Make(key)
	e, ok := m.m.Load(kp)
	if !ok {
		return nil, false
	}
	value, ok = m.loadValue(kp, e)
	return value, ok
}

// Delete is like [sync.Map.Delete].
func (m *KeyValueMap[K, V]) Delete(key *K) {
	_, _ = m.LoadAndDelete(key)
}

// Clear is like [sync.Map.Clear].
func (m *KeyValueMap[K, V]) Clear() {
	m.ensureInit()
	clearMap(m.m, m.deleteEntry)
}

// Swap is like [sync.Map.Swap].
func (m *KeyValueMap[K, V]) Swap(key *K, value *V) (previous *V, loaded bool) {
	m.ensureInit()
	kp := weak.Make(key)
	e, ok := m.m.Load(kp)
	if ok {
		v, alive := loadPointer(e.value)
		if alive && v == value {
			return v, true
		}
	}
	e = m.newEntry(key, kp, value)
	e, ok = m.m.Swap(kp, e)
	if ok {
		previous, loaded = loadPointer(e.value)
		e.stopCleanup()
	}
	m.maybeSweep(m.sweep)
	return previous, loaded
}

// LoadAndDelete is like [sync.Map.LoadAndDelete].
func (m *KeyValueMap[K, V]) LoadAndDelete(key *K) (value *V, loaded bool) {
	m.ensureInit()
	kp := weak.Make(key)
	e, ok := m.m.LoadAndDelete(kp)
	if ok {
		value, loaded = loadPointer(e.value)
		e.stopCleanup()
		m.maybeSweep(m.sweep)
	}
	return value, loaded
}

// LoadOrStore is like [sync.Map.LoadOrStore].
func (m *KeyValueMap[K, V]) LoadOrStore(key *K, value *V) (actual *V, loaded bool) {
	m.ensureInit()
	kp := weak.Make(key)
	for {
		e, ok := m.m.Load(kp)
		if ok {
			actual, loaded = loadPointer(e.value)
			if loaded {
				return actual, true
			}
		}
		ne := m.newEntry(key, kp, value)
		prev, loaded := m.m.LoadOrStore(kp, ne)
		if !loaded {
			m.maybeSweep(m.sweep)
			return value, false
		}
		ne.stopCleanup()
		m.loadValue(kp, prev)
	}
}

// CompareAndDelete is like [sync.Map.CompareAndDelete].
func (m *KeyValueMap[K, V]) CompareAndDelete(key *K, old *V) (deleted bool) {
	m.ensureInit()
	kp := weak.Make(key)
	for {
		e, ok := m.m.Load(kp)
		if !ok {
			return false
		}
		v, alive := m.loadValue(kp, e)
		if !alive {
			return false
		}
		if v != old {
			return false
		}
		if m.deleteEntry(kp, e) {
			m.maybeSweep(m.sweep)
			return true
		}
	}
}

// CompareAndSwap is like [sync.Map.CompareAndSwap].
func (m *KeyValueMap[K, V]) CompareAndSwap(key *K, oldValue, newValue *V) (swapped bool) {
	m.ensureInit()
	kp := weak.Make(key)
	for {
		e, ok := m.m.Load(kp)
		if !ok {
			return false
		}
		v, alive := m.loadValue(kp, e)
		if !alive {
			return false
		}
		if v != oldValue {
			return false
		}
		if oldValue == newValue {
			return true
		}
		ne := m.newEntry(key, kp, newValue)
		swapped = m.m.CompareAndSwap(kp, e, ne)
		if swapped {
			ne = e
		}
		ne.stopCleanup()
		if swapped {
			m.maybeSweep(m.sweep)
			return true
		}
	}
}

// Range is like [sync.Map.Range].
func (m *KeyValueMap[K, V]) Range(f func(key *K, value *V) bool) {
	m.ensureInit()
	m.m.Range(func(kp weak.Pointer[K], e keyValueMapEntry[K, V]) bool {
		key, alive := m.loadKey(kp, e)
		if !alive {
			return true
		}
		value, alive := m.loadValue(kp, e)
		if !alive {
			return true
		}
		return f(key, value)
	})
}

// All returns an iterator over all entries in the map.
// See [KeyValueMap.Range] for more details.
func (m *KeyValueMap[K, V]) All() iter.Seq2[*K, *V] {
	return m.Range
}
