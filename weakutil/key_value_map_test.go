package weakutil_test

import (
	"fmt"
	"runtime"
	"testing"
	"time"
	"weak"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/weakutil"
)

func ExampleKeyValueMap() {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{} // Must use a large key in order to trigger garbage collection reliably.
	v := &[64]byte{} // Must use a large value in order to trigger garbage collection reliably.
	m.Store(k, v)
	runtime.GC()
	fmt.Println(m.Load(k)) // Both are still valid, because there are keepalives below.
	runtime.KeepAlive(k)
	runtime.KeepAlive(v)
	runtime.GC()
	count := 0
	for range m.Range {
		count++
	}
	fmt.Println(count) // The key is not valid anymore.
	// Output:
	// &[0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0] true
	// 0
}

func TestKeyValueMapStoreReplace(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	v1 := &[64]byte{}
	m.Store(k, v1)
	v2 := &[64]byte{}
	m.Store(k, v2)
	v3, ok := m.Load(k)
	assert.True(t, ok)
	assert.Equal(t, v3, v2)
	assert.Equal(t, getKeyValueMapLen(m), 1)
	runtime.KeepAlive(k)
	runtime.KeepAlive(v1)
	runtime.KeepAlive(v2)
}

func TestKeyValueMapStoreNilKey(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	v1 := &[64]byte{}
	m.Store(nil, v1)
	v2, ok := m.Load(nil)
	assert.True(t, ok)
	assert.Equal(t, v2, v1)
	assert.Equal(t, getKeyValueMapLen(m), 1)
	runtime.KeepAlive(v1)
}

func TestKeyValueMapStoreNilValue(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	m.Store(k, nil)
	v, ok := m.Load(k)
	assert.True(t, ok)
	assert.Zero(t, v)
	assert.Equal(t, getKeyValueMapLen(m), 1)
	runtime.KeepAlive(k)
}

func TestKeyValueMapStoreNilBoth(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	m.Store(nil, nil)
	v, ok := m.Load(nil)
	assert.True(t, ok)
	assert.Zero(t, v)
	assert.Equal(t, getKeyValueMapLen(m), 1)
}

func BenchmarkKeyValueMapStoreSame(b *testing.B) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	v := &[64]byte{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Store(k, v)
		}
	})
	runtime.KeepAlive(k)
	runtime.KeepAlive(v)
}

func BenchmarkKeyValueMapStoreDifferent(b *testing.B) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var vs [2][64]byte
		for i := 0; pb.Next(); i++ {
			v := &vs[i%2]
			m.Store(k, v)
		}
	})
	runtime.KeepAlive(k)
}

func BenchmarkKeyValueMapStoreNewRandomKey(b *testing.B) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			k := &[64]byte{}
			v := &[64]byte{}
			m.Store(k, v)
		}
	})
}

func TestKeyValueMapLoadNotFound(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	v, ok := m.Load(nil)
	assert.False(t, ok)
	assert.Zero(t, v)
	assert.Equal(t, getKeyValueMapLen(m), 0)
}

func TestKeyValueMapLoadRemovedGCKey(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	v := &[64]byte{}
	func() {
		k := &[64]byte{}
		m.Store(k, v)
	}()
	runtime.GC()
	assert.Equal(t, getKeyValueMapLen(m), 0)
	runtime.KeepAlive(v)
}

func BenchmarkKeyValueMapLoad(b *testing.B) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	v := &[64]byte{}
	m.Store(k, v)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Load(k)
		}
	})
	runtime.KeepAlive(k)
	runtime.KeepAlive(v)
}

func BenchmarkKeyValueMapLoadNil(b *testing.B) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	m.Store(nil, nil)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Load(nil)
		}
	})
}

func BenchmarkKeyValueMapLoadNotFound(b *testing.B) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Load(k)
		}
	})
}

func TestKeyValueMapDelete(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	v1 := &[64]byte{}
	m.Store(k, v1)
	m.Delete(k)
	v2, ok := m.Load(k)
	assert.False(t, ok)
	assert.Zero(t, v2)
	assert.Equal(t, getKeyValueMapLen(m), 0)
	runtime.KeepAlive(k)
	runtime.KeepAlive(v1)
}

func TestKeyValueMapDeleteNotFound(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	m.Delete(k)
	assert.Equal(t, getKeyValueMapLen(m), 0)
	runtime.KeepAlive(k)
}

func BenchmarkKeyValueMapDelete(b *testing.B) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Delete(k)
		}
	})
	runtime.KeepAlive(k)
}

func TestKeyValueMapClear(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	v1 := &[64]byte{}
	m.Store(k, v1)
	m.Clear()
	v2, ok := m.Load(k)
	assert.False(t, ok)
	assert.Zero(t, v2)
	assert.Equal(t, getKeyValueMapLen(m), 0)
	runtime.KeepAlive(k)
	runtime.KeepAlive(v1)
}

func TestKeyValueMapClearEmpty(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	m.Clear()
	assert.Equal(t, getKeyValueMapLen(m), 0)
}

func BenchmarkKeyValueMapClear(b *testing.B) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Clear()
		}
	})
}

func TestKeyValueMapSwap(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	v1 := &[64]byte{}
	m.Store(k, v1)
	v2 := &[64]byte{}
	v3, loaded := m.Swap(k, v2)
	assert.True(t, loaded)
	assert.Equal(t, v3, v1)
	v4, ok := m.Load(k)
	assert.True(t, ok)
	assert.Equal(t, v4, v2)
	assert.Equal(t, getKeyValueMapLen(m), 1)
	runtime.KeepAlive(k)
	runtime.KeepAlive(v1)
	runtime.KeepAlive(v2)
}

func TestKeyValueMapSwapSame(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	v1 := &[64]byte{}
	m.Store(k, v1)
	v2, loaded := m.Swap(k, v1)
	assert.True(t, loaded)
	assert.Equal(t, v2, v1)
	v3, ok := m.Load(k)
	assert.True(t, ok)
	assert.Equal(t, v3, v1)
	assert.Equal(t, getKeyValueMapLen(m), 1)
	runtime.KeepAlive(k)
	runtime.KeepAlive(v1)
}

func TestKeyValueMapSwapNotFound(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	v1 := &[64]byte{}
	v2, loaded := m.Swap(k, v1)
	assert.False(t, loaded)
	assert.Zero(t, v2)
	v3, ok := m.Load(k)
	assert.True(t, ok)
	assert.Equal(t, v3, v1)
	assert.Equal(t, getKeyValueMapLen(m), 1)
	runtime.KeepAlive(k)
	runtime.KeepAlive(v1)
}

func TestKeyValueMapLoadAndDelete(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	v1 := &[64]byte{}
	m.Store(k, v1)
	v2, loaded := m.LoadAndDelete(k)
	assert.True(t, loaded)
	assert.Equal(t, v2, v1)
	assert.Equal(t, getKeyValueMapLen(m), 0)
	runtime.KeepAlive(k)
	runtime.KeepAlive(v1)
}

func TestKeyValueMapLoadAndDeleteNotFound(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	v, loaded := m.LoadAndDelete(k)
	assert.False(t, loaded)
	assert.Zero(t, v)
	assert.Equal(t, getKeyValueMapLen(m), 0)
	runtime.KeepAlive(k)
}

func BenchmarkKeyValueMapLoadAndDelete(b *testing.B) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.LoadAndDelete(k)
		}
	})
	runtime.KeepAlive(k)
}

func TestKeyValueMapLoadOrStore(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	v1 := &[64]byte{}
	m.Store(k, v1)
	v2 := &[64]byte{}
	v3, loaded := m.LoadOrStore(k, v2)
	assert.True(t, loaded)
	assert.Equal(t, v3, v1)
	assert.Equal(t, getKeyValueMapLen(m), 1)
	runtime.KeepAlive(k)
	runtime.KeepAlive(v1)
}

func TestKeyValueMapLoadOrStoreNotFound(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	v1 := &[64]byte{}
	v2, loaded := m.LoadOrStore(k, v1)
	assert.False(t, loaded)
	assert.Equal(t, v2, v1)
	assert.Equal(t, getKeyValueMapLen(m), 1)
	runtime.KeepAlive(k)
	runtime.KeepAlive(v1)
}

func TestKeyValueMapLoadOrStoreDeadEntry(t *testing.T) {
	disableGC(t)
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	v := &[64]byte{}
	runtime.SetFinalizer(v, func(*[64]byte) {}) // keeps the cleanup from running until a second GC
	m.Store(k, v)
	runtime.KeepAlive(k) // keep k alive through Store
	runtime.KeepAlive(v) // keep v alive through Store
	runtime.GC()         // v is unreachable: weak handle cleared (dead entry), cleanup kept until a second GC
	done := make(chan struct{})
	go func() {
		m.LoadOrStore(k, &[64]byte{})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("KeyValueMap.LoadOrStore hung on an entry whose value was collected but not yet cleaned up")
	}
	runtime.KeepAlive(k)
}

func BenchmarkKeyValueMapLoadOrStore(b *testing.B) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	v := &[64]byte{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.LoadOrStore(k, v)
		}
	})
	runtime.KeepAlive(k)
	runtime.KeepAlive(v)
}

func TestKeyValueMapCompareAndDelete(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	v := &[64]byte{}
	m.Store(k, v)
	deleted := m.CompareAndDelete(k, v)
	assert.True(t, deleted)
	assert.Equal(t, getKeyValueMapLen(m), 0)
	runtime.KeepAlive(k)
	runtime.KeepAlive(v)
}

func TestKeyValueMapCompareAndDeleteNotEqual(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	v1 := &[64]byte{}
	m.Store(k, v1)
	v2 := &[64]byte{}
	deleted := m.CompareAndDelete(k, v2)
	assert.False(t, deleted)
	assert.Equal(t, getKeyValueMapLen(m), 1)
	runtime.KeepAlive(k)
	runtime.KeepAlive(v1)
	runtime.KeepAlive(v2)
}

func TestKeyValueMapCompareAndDeleteNotFound(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	v := &[64]byte{}
	deleted := m.CompareAndDelete(k, v)
	assert.False(t, deleted)
	assert.Equal(t, getKeyValueMapLen(m), 0)
	runtime.KeepAlive(k)
	runtime.KeepAlive(v)
}

func BenchmarkKeyValueMapCompareAndDelete(b *testing.B) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	v := &[64]byte{}
	m.Store(k, v)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.CompareAndDelete(k, v)
		}
	})
	runtime.KeepAlive(k)
	runtime.KeepAlive(v)
}

func TestKeyValueMapCompareAndSwap(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	v1 := &[64]byte{}
	m.Store(k, v1)
	v2 := &[64]byte{}
	swapped := m.CompareAndSwap(k, v1, v2)
	assert.True(t, swapped)
	v3, ok := m.Load(k)
	assert.True(t, ok)
	assert.Equal(t, v3, v2)
	assert.Equal(t, getKeyValueMapLen(m), 1)
	runtime.KeepAlive(k)
	runtime.KeepAlive(v1)
	runtime.KeepAlive(v2)
}

func TestKeyValueMapCompareAndSwapNotFound(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	v1 := &[64]byte{}
	v2 := &[64]byte{}
	swapped := m.CompareAndSwap(k, v2, v1)
	assert.False(t, swapped)
	_, ok := m.Load(k)
	assert.False(t, ok)
	assert.Equal(t, getKeyValueMapLen(m), 0)
	runtime.KeepAlive(k)
	runtime.KeepAlive(v1)
	runtime.KeepAlive(v2)
}

func TestKeyValueMapCompareAndSwapNotEqual(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	v1 := &[64]byte{}
	m.Store(k, v1)
	v2 := &[64]byte{}
	swapped := m.CompareAndSwap(k, v2, v1)
	assert.False(t, swapped)
	v3, ok := m.Load(k)
	assert.True(t, ok)
	assert.Equal(t, v3, v1)
	assert.Equal(t, getKeyValueMapLen(m), 1)
	runtime.KeepAlive(k)
	runtime.KeepAlive(v1)
	runtime.KeepAlive(v2)
}

func TestKeyValueMapCompareAndSwapSame(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	v1 := &[64]byte{}
	m.Store(k, v1)
	swapped := m.CompareAndSwap(k, v1, v1)
	assert.True(t, swapped)
	v2, ok := m.Load(k)
	assert.True(t, ok)
	assert.Equal(t, v2, v1)
	assert.Equal(t, getKeyValueMapLen(m), 1)
	runtime.KeepAlive(k)
	runtime.KeepAlive(v1)
}

func BenchmarkKeyValueMapCompareAndSwapNotEqual(b *testing.B) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	v1 := &[64]byte{}
	m.Store(k, v1)
	v2 := &[64]byte{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.CompareAndSwap(k, v2, v2)
		}
	})
	runtime.KeepAlive(k)
	runtime.KeepAlive(v1)
}

func BenchmarkKeyValueMapCompareAndSwapSame(b *testing.B) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	v := &[64]byte{}
	m.Store(k, v)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.CompareAndSwap(k, v, v)
		}
	})
	runtime.KeepAlive(k)
	runtime.KeepAlive(v)
}

func TestKeyValueMapRange(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k := &[64]byte{}
	v1 := &[64]byte{}
	m.Store(k, v1)
	found := false
	for k2, v2 := range m.All() {
		assert.Equal(t, k2, k)
		assert.Equal(t, v2, v1)
		found = true
	}
	assert.True(t, found)
	assert.Equal(t, getKeyValueMapLen(m), 1)
	runtime.KeepAlive(k)
	runtime.KeepAlive(v1)
}

func TestKeyValueMapRangeNil(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	m.Store(nil, nil)
	found := false
	for k, v := range m.All() {
		assert.Zero(t, k)
		assert.Zero(t, v)
		found = true
	}
	assert.True(t, found)
}

func TestKeyValueMapRangeInterrupt(t *testing.T) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	k1 := &[64]byte{}
	k2 := &[64]byte{}
	v1 := &[64]byte{}
	m.Store(k1, v1)
	m.Store(k2, v1)
	for range m.All() {
		break
	}
	runtime.KeepAlive(k1)
	runtime.KeepAlive(k2)
	runtime.KeepAlive(v1)
}

func keyValueMapStoreDeadValue[K any, V any](m *KeyValueMap[K, V], key *K) (kp weak.Pointer[K]) {
	kp = weak.Make(key)
	v := new(V)
	runtime.SetFinalizer(v, func(*V) {}) // keeps the cleanup from running until a second GC
	m.Store(key, v)
	runtime.KeepAlive(key) // keep key alive through Store
	runtime.KeepAlive(v)   // keep v alive through Store
	return kp
}

func assertKeyValueMapDeadEntry[K any, V any](tb testing.TB, m *KeyValueMap[K, V], kp weak.Pointer[K]) {
	tb.Helper()
	present, alive := KeyValueMapRawEntry(m, kp)
	assert.True(tb, present)
	assert.False(tb, alive)
}

func assertKeyValueMapNoEntry[K any, V any](tb testing.TB, m *KeyValueMap[K, V], kp weak.Pointer[K]) {
	tb.Helper()
	present, _ := KeyValueMapRawEntry(m, kp)
	assert.False(tb, present)
}

func TestKeyValueMapLoadEvictsDeadValue(t *testing.T) {
	disableGC(t)
	m := new(KeyValueMap[int, [64]byte])
	k := new(int)
	kp := keyValueMapStoreDeadValue(m, k)
	runtime.GC() // v is unreachable: dead value, cleanups kept until a second GC
	assertKeyValueMapDeadEntry(t, m, kp)
	val, ok := m.Load(k)
	assert.False(t, ok)
	assert.Zero(t, val)
	assertKeyValueMapNoEntry(t, m, kp)
}

func TestKeyValueMapRangeEvictsDeadValue(t *testing.T) {
	disableGC(t)
	m := new(KeyValueMap[int, [64]byte])
	k := new(int)
	kp := keyValueMapStoreDeadValue(m, k)
	runtime.GC() // v is unreachable: dead value, cleanups kept until a second GC
	assertKeyValueMapDeadEntry(t, m, kp)
	runtime.KeepAlive(k) // k must stay alive through the GCs, or its own cleanup removes the entry
	seen := false
	for range m.All() {
		seen = true
	}
	assert.False(t, seen)
	assertKeyValueMapNoEntry(t, m, kp)
}

func TestKeyValueMapCompareAndDeleteEvictsDeadValue(t *testing.T) {
	disableGC(t)
	m := new(KeyValueMap[int, [64]byte])
	k := new(int)
	kp := keyValueMapStoreDeadValue(m, k)
	runtime.GC() // v is unreachable: dead value, cleanups kept until a second GC
	assertKeyValueMapDeadEntry(t, m, kp)
	assert.False(t, m.CompareAndDelete(k, &[64]byte{}))
	assertKeyValueMapNoEntry(t, m, kp)
}

func TestKeyValueMapCompareAndSwapEvictsDeadValue(t *testing.T) {
	disableGC(t)
	m := new(KeyValueMap[int, [64]byte])
	k := new(int)
	kp := keyValueMapStoreDeadValue(m, k)
	runtime.GC() // v is unreachable: dead value, cleanups kept until a second GC
	assertKeyValueMapDeadEntry(t, m, kp)
	assert.False(t, m.CompareAndSwap(k, &[64]byte{}, &[64]byte{}))
	assertKeyValueMapNoEntry(t, m, kp)
}

func BenchmarkKeyValueMapRange(b *testing.B) {
	m := new(KeyValueMap[[64]byte, [64]byte])
	var ks [10]*[64]byte
	v := &[64]byte{}
	for i := range 10 {
		k := &[64]byte{}
		ks[i] = k
		m.Store(k, v)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for range m.All() {
			}
		}
	})
	runtime.KeepAlive(ks)
	runtime.KeepAlive(v)
}

func getKeyValueMapLen[K any, V any](m *KeyValueMap[K, V]) int {
	count := 0
	for range m.Range {
		count++
	}
	return count
}
