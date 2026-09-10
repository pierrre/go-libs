package weakutil

import (
	"iter"
	"runtime"
	"sync"
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
	m                    syncutil.Map[weak.Pointer[K], keyValueMapValue[K, V]]
	keyCleanupFunc       func(weak.Pointer[K])
	keyCleanupFuncOnce   sync.Once
	valueCleanupFunc     func(keyValueMapCleanupArg[K, V])
	valueCleanupFuncOnce sync.Once
}

type keyValueMapValue[K any, V any] struct {
	value        weak.Pointer[V]
	keyCleanup   runtime.Cleanup
	valueCleanup runtime.Cleanup
}

func (mv *keyValueMapValue[K, V]) stopCleanup() {
	mv.keyCleanup.Stop()
	mv.valueCleanup.Stop()
}

type keyValueMapCleanupArg[K any, V any] struct {
	key   weak.Pointer[K]
	value weak.Pointer[V]
}

func (m *KeyValueMap[K, V]) getKeyCleanupFunc() func(weak.Pointer[K]) {
	m.keyCleanupFuncOnce.Do(func() {
		m.keyCleanupFunc = m.keyCleanup
	})
	return m.keyCleanupFunc
}

func (m *KeyValueMap[K, V]) getValueCleanupFunc() func(keyValueMapCleanupArg[K, V]) {
	m.valueCleanupFuncOnce.Do(func() {
		m.valueCleanupFunc = m.valueCleanup
	})
	return m.valueCleanupFunc
}

func (m *KeyValueMap[K, V]) newValue(key *K, kp weak.Pointer[K], value *V) (mv keyValueMapValue[K, V]) {
	if key != nil {
		mv.keyCleanup = runtime.AddCleanup(key, m.getKeyCleanupFunc(), kp)
	}
	if value != nil {
		vp := weak.Make(value)
		mv.value = vp
		mv.valueCleanup = runtime.AddCleanup(value, m.getValueCleanupFunc(), keyValueMapCleanupArg[K, V]{
			key:   kp,
			value: vp,
		})
	}
	return mv
}

func (m *KeyValueMap[K, V]) deleteValue(kp weak.Pointer[K], mv keyValueMapValue[K, V]) (deleted bool) {
	mv.stopCleanup()
	return m.m.CompareAndDelete(kp, mv)
}

func (m *KeyValueMap[K, V]) loadValue(kp weak.Pointer[K], mv keyValueMapValue[K, V]) (value *V, alive bool) {
	value, alive = loadPointer(mv.value)
	if !alive {
		m.deleteValue(kp, mv)
	}
	return value, alive
}

func (m *KeyValueMap[K, V]) loadKey(kp weak.Pointer[K], mv keyValueMapValue[K, V]) (key *K, alive bool) {
	key, alive = loadPointer(kp)
	if !alive {
		m.deleteValue(kp, mv)
	}
	return key, alive
}

func (m *KeyValueMap[K, V]) keyCleanup(kp weak.Pointer[K]) {
	mv, ok := m.m.LoadAndDelete(kp)
	if ok {
		_, alive := loadPointer(mv.value)
		if alive {
			mv.valueCleanup.Stop()
		}
	}
}

func (m *KeyValueMap[K, V]) valueCleanup(mc keyValueMapCleanupArg[K, V]) {
	mv, ok := m.m.Load(mc.key)
	if ok && mv.value == mc.value {
		m.m.CompareAndDelete(mc.key, mv)
		_, alive := loadPointer(mc.key)
		if alive {
			mv.keyCleanup.Stop()
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
	mv, ok := m.m.Load(kp)
	if !ok {
		return nil, false
	}
	value, ok = m.loadValue(kp, mv)
	return value, ok
}

// Delete is like [sync.Map.Delete].
func (m *KeyValueMap[K, V]) Delete(key *K) {
	_, _ = m.LoadAndDelete(key)
}

// Clear is like [sync.Map.Clear].
func (m *KeyValueMap[K, V]) Clear() {
	for {
		var count int64
		m.m.Range(func(kp weak.Pointer[K], mv keyValueMapValue[K, V]) bool {
			count++
			m.deleteValue(kp, mv)
			return true
		})
		if count == 0 {
			return
		}
	}
}

// Swap is like [sync.Map.Swap].
func (m *KeyValueMap[K, V]) Swap(key *K, value *V) (previous *V, loaded bool) {
	kp := weak.Make(key)
	mv, ok := m.m.Load(kp)
	if ok {
		v, alive := loadPointer(mv.value)
		if alive && v == value {
			return v, true
		}
	}
	mv = m.newValue(key, kp, value)
	mv, ok = m.m.Swap(kp, mv)
	if ok {
		previous, loaded = loadPointer(mv.value)
		mv.stopCleanup()
	}
	return previous, loaded
}

// LoadAndDelete is like [sync.Map.LoadAndDelete].
func (m *KeyValueMap[K, V]) LoadAndDelete(key *K) (value *V, loaded bool) {
	kp := weak.Make(key)
	mv, ok := m.m.LoadAndDelete(kp)
	if ok {
		value, loaded = loadPointer(mv.value)
		mv.stopCleanup()
	}
	return value, loaded
}

// LoadOrStore is like [sync.Map.LoadOrStore].
func (m *KeyValueMap[K, V]) LoadOrStore(key *K, value *V) (actual *V, loaded bool) {
	kp := weak.Make(key)
	for {
		mv, ok := m.m.Load(kp)
		if ok {
			actual, loaded = loadPointer(mv.value)
			if loaded {
				return actual, true
			}
		}
		newMv := m.newValue(key, kp, value)
		prev, loaded := m.m.LoadOrStore(kp, newMv)
		if !loaded {
			return value, false
		}
		newMv.stopCleanup()
		m.loadValue(kp, prev)
	}
}

// CompareAndDelete is like [sync.Map.CompareAndDelete].
func (m *KeyValueMap[K, V]) CompareAndDelete(key *K, old *V) (deleted bool) {
	kp := weak.Make(key)
	for {
		mv, ok := m.m.Load(kp)
		if !ok {
			return false
		}
		v, alive := m.loadValue(kp, mv)
		if !alive {
			return false
		}
		if v != old {
			return false
		}
		if m.deleteValue(kp, mv) {
			return true
		}
	}
}

// CompareAndSwap is like [sync.Map.CompareAndSwap].
func (m *KeyValueMap[K, V]) CompareAndSwap(key *K, oldValue, newValue *V) (swapped bool) {
	kp := weak.Make(key)
	for {
		mv, ok := m.m.Load(kp)
		if !ok {
			return false
		}
		v, alive := m.loadValue(kp, mv)
		if !alive {
			return false
		}
		if v != oldValue {
			return false
		}
		if oldValue == newValue {
			return true
		}
		newMv := m.newValue(key, kp, newValue)
		swapped = m.m.CompareAndSwap(kp, mv, newMv)
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
func (m *KeyValueMap[K, V]) Range(f func(key *K, value *V) bool) {
	m.m.Range(func(kp weak.Pointer[K], mv keyValueMapValue[K, V]) bool {
		key, alive := m.loadKey(kp, mv)
		if !alive {
			return true
		}
		value, alive := m.loadValue(kp, mv)
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
