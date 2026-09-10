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
// An entry whose value has been collected by the garbage collector is treated as absent until its cleanup evicts it.
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

func (m *KeyValueMap[K, V]) keyCleanup(kp weak.Pointer[K]) {
	mv, ok := m.m.LoadAndDelete(kp)
	if ok {
		_, ok = loadPointer(mv.value)
		if ok {
			mv.valueCleanup.Stop()
		}
	}
}

func (m *KeyValueMap[K, V]) valueCleanup(mc keyValueMapCleanupArg[K, V]) {
	mv, ok := m.m.Load(mc.key)
	if ok && mv.value == mc.value {
		m.m.CompareAndDelete(mc.key, mv)
		_, ok = loadPointer(mc.key)
		if ok {
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
	return loadPointer(mv.value)
}

// Delete is like [sync.Map.Delete].
func (m *KeyValueMap[K, V]) Delete(key *K) {
	_, _ = m.LoadAndDelete(key)
}

// Clear is like [sync.Map.Clear].
func (m *KeyValueMap[K, V]) Clear() {
	m.m.Range(func(kp weak.Pointer[K], mv keyValueMapValue[K, V]) bool {
		mv.stopCleanup()
		return true
	})
	m.m.Clear()
}

// Swap is like [sync.Map.Swap].
func (m *KeyValueMap[K, V]) Swap(key *K, value *V) (previous *V, loaded bool) {
	kp := weak.Make(key)
	old, ok := m.m.Load(kp)
	if ok {
		v, alive := loadPointer(old.value)
		if alive && v == value {
			return v, true
		}
	}
	mv := m.newValue(key, kp, value)
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
		old, ok := m.m.Load(kp)
		if ok {
			v, alive := loadPointer(old.value)
			if alive {
				return v, true
			}
		}
		mv := m.newValue(key, kp, value)
		prev, loaded := m.m.LoadOrStore(kp, mv)
		if !loaded {
			return value, false
		}
		mv.stopCleanup()
		_, ok = loadPointer(prev.value)
		if !ok {
			if m.m.CompareAndDelete(kp, prev) {
				prev.stopCleanup()
			}
		}
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
		v, ok := loadPointer(mv.value)
		if !ok || v != old {
			return false
		}
		deleted = m.m.CompareAndDelete(kp, mv)
		if deleted {
			mv.stopCleanup()
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
		v, ok := loadPointer(mv.value)
		if !ok || v != oldValue {
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
		key, ok := loadPointer(kp)
		if !ok {
			return true
		}
		value, ok := loadPointer(mv.value)
		if !ok {
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
