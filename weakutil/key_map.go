package weakutil

import (
	"iter"
	"runtime"
	"weak"

	"github.com/pierrre/go-libs/syncutil"
)

// KeyMap is a map that automatically evicts entries when the key is no longer reachable.
// It is safe for concurrent use.
// The zero value is ready to use.
// If a value is set with a nil key, it is never evicted.
//
// It implements the same methods as [sync.Map].
type KeyMap[K comparable, V any] struct {
	m syncutil.Map[weak.Pointer[K], keyMapValue[V]]
}

type keyMapValue[V any] struct {
	value   V
	cleanup runtime.Cleanup
}

func (m *KeyMap[K, V]) newValue(key *K, value V) (kp weak.Pointer[K], mv keyMapValue[V]) {
	mv.value = value
	if key != nil {
		kp = weak.Make(key)
		mv.cleanup = runtime.AddCleanup(key, m.cleanup, kp)
	}
	return kp, mv
}

func (m *KeyMap[K, V]) cleanup(kp weak.Pointer[K]) {
	m.m.Delete(kp)
}

// Store is like [sync.Map.Store].
func (m *KeyMap[K, V]) Store(key *K, value V) {
	kp, mv := m.newValue(key, value)
	mv, ok := m.m.Swap(kp, mv)
	if ok {
		mv.cleanup.Stop()
	}
}

// Load is like [sync.Map.Load].
func (m *KeyMap[K, V]) Load(key *K) (value V, ok bool) {
	kp := weak.Make(key)
	mv, ok := m.m.Load(kp)
	return mv.value, ok
}

// Delete is like [sync.Map.Delete].
func (m *KeyMap[K, V]) Delete(key *K) {
	kp := weak.Make(key)
	mv, ok := m.m.LoadAndDelete(kp)
	if ok {
		mv.cleanup.Stop()
	}
}

// Clear is like [sync.Map.Clear].
func (m *KeyMap[K, V]) Clear() {
	m.m.Range(func(kp weak.Pointer[K], mv keyMapValue[V]) bool {
		m.m.CompareAndDelete(kp, mv)
		mv.cleanup.Stop()
		return true
	})
}

// Swap is like [sync.Map.Swap].
func (m *KeyMap[K, V]) Swap(key *K, value V) (previous V, loaded bool) {
	kp, mv := m.newValue(key, value)
	mv, loaded = m.m.Swap(kp, mv)
	if loaded {
		mv.cleanup.Stop()
	}
	return mv.value, loaded
}

// LoadAndDelete is like [sync.Map.LoadAndDelete].
func (m *KeyMap[K, V]) LoadAndDelete(key *K) (value V, loaded bool) {
	kp := weak.Make(key)
	mv, loaded := m.m.LoadAndDelete(kp)
	if loaded {
		mv.cleanup.Stop()
	}
	return mv.value, loaded
}

// LoadOrStore is like [sync.Map.LoadOrStore].
func (m *KeyMap[K, V]) LoadOrStore(key *K, value V) (actual V, loaded bool) {
	var kp weak.Pointer[K]
	var mv keyMapValue[V]
	for {
		mv.cleanup.Stop()
		actual, loaded = m.Load(key)
		if loaded {
			return actual, true
		}
		kp, mv = m.newValue(key, value)
		_, loaded = m.m.LoadOrStore(kp, mv)
		if !loaded {
			return value, false
		}
	}
}

// CompareAndDelete is like [sync.Map.CompareAndDelete].
func (m *KeyMap[K, V]) CompareAndDelete(key *K, old V) (deleted bool) {
	kp := weak.Make(key)
	for {
		mv, ok := m.m.Load(kp)
		if !ok {
			return false
		}
		if any(mv.value) != any(old) {
			return false
		}
		deleted = m.m.CompareAndDelete(kp, mv)
		if deleted {
			mv.cleanup.Stop()
			return true
		}
	}
}

// CompareAndSwap is like [sync.Map.CompareAndSwap].
func (m *KeyMap[K, V]) CompareAndSwap(key *K, oldValue, newValue V) (swapped bool) {
	kp := weak.Make(key)
	for {
		mv, ok := m.m.Load(kp)
		if !ok {
			return false
		}
		if any(mv.value) != any(oldValue) {
			return false
		}
		_, newMv := m.newValue(key, newValue)
		swapped = m.m.CompareAndSwap(kp, mv, newMv)
		if swapped {
			newMv = mv
		}
		newMv.cleanup.Stop()
		if swapped {
			return true
		}
	}
}

// Range is like [sync.Map.Range].
func (m *KeyMap[K, V]) Range(f func(key *K, value V) bool) {
	m.m.Range(func(kp weak.Pointer[K], mv keyMapValue[V]) bool {
		key, ok := loadPointer(kp)
		return !ok || f(key, mv.value)
	})
}

// All returns an iterator over all entries in the map.
// See [KeyMap.Range] for more details.
func (m *KeyMap[K, V]) All() iter.Seq2[*K, V] {
	return m.Range
}
