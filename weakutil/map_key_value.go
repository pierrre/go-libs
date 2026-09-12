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
//
// It implements the same methods as [sync.Map].
type KeyValueMap[K any, V any] struct {
	m                syncutil.Map[weak.Pointer[K], keyValueMapEntry[K, V]]
	keyCleanupFunc   lazyValue[func(weak.Pointer[K])]
	valueCleanupFunc lazyValue[func(keyValueMapCleanupArg[K, V])]
}

type keyValueMapEntry[K any, V any] struct {
	value        weak.Pointer[V]
	keyCleanup   runtime.Cleanup
	valueCleanup runtime.Cleanup
}

func (e *keyValueMapEntry[K, V]) stopCleanup() {
	e.keyCleanup.Stop()
	e.valueCleanup.Stop()
}

type keyValueMapCleanupArg[K any, V any] struct {
	key   weak.Pointer[K]
	value weak.Pointer[V]
}

func (m *KeyValueMap[K, V]) getKeyCleanupFunc() func(weak.Pointer[K]) {
	return m.keyCleanupFunc.get(func() func(weak.Pointer[K]) {
		return m.keyCleanup
	})
}

func (m *KeyValueMap[K, V]) getValueCleanupFunc() func(keyValueMapCleanupArg[K, V]) {
	return m.valueCleanupFunc.get(func() func(keyValueMapCleanupArg[K, V]) {
		return m.valueCleanup
	})
}

func (m *KeyValueMap[K, V]) newEntry(key *K, kp weak.Pointer[K], value *V) (e keyValueMapEntry[K, V]) {
	if key != nil {
		e.keyCleanup = runtime.AddCleanup(key, m.getKeyCleanupFunc(), kp)
	}
	if value != nil {
		vp := weak.Make(value)
		e.value = vp
		e.valueCleanup = runtime.AddCleanup(value, m.getValueCleanupFunc(), keyValueMapCleanupArg[K, V]{
			key:   kp,
			value: vp,
		})
	}
	return e
}

func (m *KeyValueMap[K, V]) deleteEntry(kp weak.Pointer[K], e keyValueMapEntry[K, V]) (deleted bool) {
	e.stopCleanup()
	return m.m.CompareAndDelete(kp, e)
}

func (m *KeyValueMap[K, V]) loadValue(kp weak.Pointer[K], e keyValueMapEntry[K, V]) (value *V, alive bool) {
	value, alive = loadPointer(e.value)
	if !alive {
		m.deleteEntry(kp, e)
	}
	return value, alive
}

func (m *KeyValueMap[K, V]) loadKey(kp weak.Pointer[K], e keyValueMapEntry[K, V]) (key *K, alive bool) {
	key, alive = loadPointer(kp)
	if !alive {
		m.deleteEntry(kp, e)
	}
	return key, alive
}

func (m *KeyValueMap[K, V]) keyCleanup(kp weak.Pointer[K]) {
	e, ok := m.m.LoadAndDelete(kp)
	if ok {
		_, alive := loadPointer(e.value)
		if alive {
			e.valueCleanup.Stop()
		}
	}
}

func (m *KeyValueMap[K, V]) valueCleanup(mc keyValueMapCleanupArg[K, V]) {
	e, ok := m.m.Load(mc.key)
	if ok && e.value == mc.value {
		m.m.CompareAndDelete(mc.key, e)
		_, alive := loadPointer(mc.key)
		if alive {
			e.keyCleanup.Stop()
		}
	}
}

// Store is like [sync.Map.Store].
func (m *KeyValueMap[K, V]) Store(key *K, value *V) {
	_, _ = m.Swap(key, value)
}

// Load is like [sync.Map.Load].
func (m *KeyValueMap[K, V]) Load(key *K) (value *V, ok bool) {
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
	clearMap(&m.m, m.deleteEntry)
}

// Swap is like [sync.Map.Swap].
func (m *KeyValueMap[K, V]) Swap(key *K, value *V) (previous *V, loaded bool) {
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
	return previous, loaded
}

// LoadAndDelete is like [sync.Map.LoadAndDelete].
func (m *KeyValueMap[K, V]) LoadAndDelete(key *K) (value *V, loaded bool) {
	kp := weak.Make(key)
	e, ok := m.m.LoadAndDelete(kp)
	if ok {
		value, loaded = loadPointer(e.value)
		e.stopCleanup()
	}
	return value, loaded
}

// LoadOrStore is like [sync.Map.LoadOrStore].
func (m *KeyValueMap[K, V]) LoadOrStore(key *K, value *V) (actual *V, loaded bool) {
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
			return value, false
		}
		ne.stopCleanup()
		m.loadValue(kp, prev)
	}
}

// CompareAndDelete is like [sync.Map.CompareAndDelete].
func (m *KeyValueMap[K, V]) CompareAndDelete(key *K, old *V) (deleted bool) {
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
			return true
		}
	}
}

// CompareAndSwap is like [sync.Map.CompareAndSwap].
func (m *KeyValueMap[K, V]) CompareAndSwap(key *K, oldValue, newValue *V) (swapped bool) {
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
			return true
		}
	}
}

// Range is like [sync.Map.Range].
func (m *KeyValueMap[K, V]) Range(f func(key *K, value *V) bool) {
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
