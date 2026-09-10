package weakutil_test

import (
	"fmt"
	"math/rand/v2"
	"runtime"
	"testing"
	"time"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/weakutil"
)

func ExampleValueMap() {
	m := new(ValueMap[string, [64]byte])
	v := &[64]byte{} // Must use a large value in order to trigger garbage collection reliably.
	m.Store("test", v)
	runtime.GC()
	fmt.Println(m.Load("test")) // The pointer is still valid, because there is a keepalive below.
	runtime.KeepAlive(v)
	runtime.GC()
	fmt.Println(m.Load("test")) // The pointer is not valid anymore.
	// Output:
	// &[0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0] true
	// <nil> false
}

func TestValueMapStoreLoad(t *testing.T) {
	m := new(ValueMap[string, [64]byte]) // Must use a large value in order to trigger garbage collection reliably.
	v1 := &[64]byte{}
	m.Store("test", v1)
	m.Store("test", v1)
	v2, ok := m.Load("test")
	assert.True(t, ok)
	assert.Equal(t, v2, v1)
	assert.Equal(t, getValueMapLen(m), 1)
	runtime.KeepAlive(v1)
}

func TestValueMapStoreReplace(t *testing.T) {
	m := new(ValueMap[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	v2 := &[64]byte{}
	m.Store("test", v2)
	v3, ok := m.Load("test")
	assert.True(t, ok)
	assert.Equal(t, v2, v3)
	assert.Equal(t, getValueMapLen(m), 1)
	runtime.KeepAlive(v1)
	runtime.KeepAlive(v2)
}

func BenchmarkValueMapStoreSame(b *testing.B) {
	m := new(ValueMap[string, [64]byte])
	v := &[64]byte{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Store("test", v)
		}
	})
}

func BenchmarkValueMapStoreDifferent(b *testing.B) {
	m := new(ValueMap[string, [64]byte])
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var vs [2][64]byte
		for i := 0; pb.Next(); i++ {
			v := &vs[i%2]
			m.Store("test", v)
		}
	})
}

func BenchmarkValueMapStoreNewRandomKey(b *testing.B) {
	m := new(ValueMap[int64, [64]byte])
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Store(rand.Int64(), &[64]byte{})
		}
	})
}

func TestValueMapLoadNil(t *testing.T) {
	m := new(ValueMap[string, [64]byte])
	m.Store("test", nil)
	v, ok := m.Load("test")
	assert.True(t, ok)
	assert.Zero(t, v)
	assert.Equal(t, getValueMapLen(m), 1)
}

func TestValueMapLoadNotFound(t *testing.T) {
	m := new(ValueMap[string, [64]byte])
	v, ok := m.Load("test")
	assert.False(t, ok)
	assert.Zero(t, v)
	assert.Equal(t, getValueMapLen(m), 0)
}

func TestValueMapLoadRemovedGC(t *testing.T) {
	m := new(ValueMap[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	runtime.GC()
	v2, ok := m.Load("test")
	assert.False(t, ok)
	assert.Zero(t, v2)
	assert.Equal(t, getValueMapLen(m), 0)
}

func BenchmarkValueMapLoad(b *testing.B) {
	m := new(ValueMap[string, [64]byte])
	v := &[64]byte{}
	m.Store("test", v)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Load("test")
		}
	})
}

func BenchmarkValueMapLoadNil(b *testing.B) {
	m := new(ValueMap[string, [64]byte])
	m.Store("test", nil)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Load("test")
		}
	})
}

func BenchmarkValueMapLoadNotFound(b *testing.B) {
	m := new(ValueMap[string, [64]byte])
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Load("test")
		}
	})
}

func TestValueMapDelete(t *testing.T) {
	m := new(ValueMap[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	m.Delete("test")
	v2, ok := m.Load("test")
	assert.False(t, ok)
	assert.Zero(t, v2)
	assert.Equal(t, getValueMapLen(m), 0)
	runtime.KeepAlive(v1)
}

func TestValueMapDeleteNotFound(t *testing.T) {
	m := new(ValueMap[string, [64]byte])
	m.Delete("test")
	assert.Equal(t, getValueMapLen(m), 0)
}

func BenchmarkValueMapDelete(b *testing.B) {
	m := new(ValueMap[string, [64]byte])
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Delete("test")
		}
	})
}

func TestValueMapClear(t *testing.T) {
	m := new(ValueMap[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	m.Clear()
	v2, ok := m.Load("test")
	assert.False(t, ok)
	assert.Zero(t, v2)
	assert.Equal(t, getValueMapLen(m), 0)
	runtime.KeepAlive(v1)
}

func TestValueMapClearEmpty(t *testing.T) {
	m := new(ValueMap[string, [64]byte])
	m.Clear()
	assert.Equal(t, getValueMapLen(m), 0)
}

func BenchmarkValueMapClear(b *testing.B) {
	m := new(ValueMap[string, [64]byte])
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Clear()
		}
	})
}

func TestValueMapSwap(t *testing.T) {
	m := new(ValueMap[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	v2 := &[64]byte{}
	v3, loaded := m.Swap("test", v2)
	assert.True(t, loaded)
	assert.Equal(t, v3, v1)
	v4, ok := m.Load("test")
	assert.True(t, ok)
	assert.Equal(t, v4, v2)
	assert.Equal(t, getValueMapLen(m), 1)
	runtime.KeepAlive(v1)
	runtime.KeepAlive(v2)
}

func TestValueMapSwapSame(t *testing.T) {
	m := new(ValueMap[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	v2, loaded := m.Swap("test", v1)
	assert.True(t, loaded)
	assert.Equal(t, v2, v1)
	v3, ok := m.Load("test")
	assert.True(t, ok)
	assert.Equal(t, v3, v1)
	assert.Equal(t, getValueMapLen(m), 1)
	runtime.KeepAlive(v1)
}

func TestValueMapSwapNotFound(t *testing.T) {
	m := new(ValueMap[string, [64]byte])
	v1 := &[64]byte{}
	v2, loaded := m.Swap("test", v1)
	assert.False(t, loaded)
	assert.Zero(t, v2)
	v3, ok := m.Load("test")
	assert.True(t, ok)
	assert.Equal(t, v3, v1)
	assert.Equal(t, getValueMapLen(m), 1)
	runtime.KeepAlive(v1)
}

func BenchmarkValueMapSwapSame(b *testing.B) {
	m := new(ValueMap[string, [64]byte])
	v := &[64]byte{}
	m.Store("test", v)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Swap("test", v)
		}
	})
}

func BenchmarkValueMapSwapDifferent(b *testing.B) {
	m := new(ValueMap[string, [64]byte])
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var vs [2][64]byte
		for i := 0; pb.Next(); i++ {
			v := &vs[i%2]
			m.Swap("test", v)
		}
	})
}

func TestValueMapLoadAndDelete(t *testing.T) {
	m := new(ValueMap[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	v2, loaded := m.LoadAndDelete("test")
	assert.True(t, loaded)
	assert.Equal(t, v1, v2)
	assert.Equal(t, getValueMapLen(m), 0)
}

func TestValueMapLoadAndDeleteNotFound(t *testing.T) {
	m := new(ValueMap[string, [64]byte])
	v, loaded := m.LoadAndDelete("test")
	assert.False(t, loaded)
	assert.Zero(t, v)
	assert.Equal(t, getValueMapLen(m), 0)
}

func BenchmarkValueMapLoadAndDelete(b *testing.B) {
	m := new(ValueMap[string, [64]byte])
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.LoadAndDelete("test")
		}
	})
}

func TestValueMapLoadOrStore(t *testing.T) {
	m := new(ValueMap[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	v2 := &[64]byte{}
	v3, loaded := m.LoadOrStore("test", v2)
	assert.True(t, loaded)
	assert.Equal(t, v3, v1)
	assert.Equal(t, getValueMapLen(m), 1)
	runtime.KeepAlive(v1)
}

func TestValueMapLoadOrStoreNotFound(t *testing.T) {
	m := new(ValueMap[string, [64]byte])
	v1 := &[64]byte{}
	v2, loaded := m.LoadOrStore("test", v1)
	assert.False(t, loaded)
	assert.Equal(t, v2, v1)
	assert.Equal(t, getValueMapLen(m), 1)
	runtime.KeepAlive(v1)
}

func TestValueMapLoadOrStoreDeadEntry(t *testing.T) {
	disableGC(t)
	m := new(ValueMap[string, [64]byte])
	v := &[64]byte{}
	runtime.SetFinalizer(v, func(*[64]byte) {}) // keeps the cleanup from running until a second GC
	m.Store("test", v)
	runtime.KeepAlive(v) // keep v alive through Store
	runtime.GC()         // v is unreachable: weak handle cleared (dead entry), cleanup kept until a second GC
	done := make(chan struct{})
	go func() {
		m.LoadOrStore("test", &[64]byte{})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("ValueMap.LoadOrStore hung on an entry whose value was collected but not yet cleaned up")
	}
}

func BenchmarkValueMapLoadOrStore(b *testing.B) {
	m := new(ValueMap[string, [64]byte])
	v := &[64]byte{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.LoadOrStore("test", v)
		}
	})
	runtime.KeepAlive(v)
}

func TestValueMapCompareAndDelete(t *testing.T) {
	m := new(ValueMap[string, [64]byte])
	v := &[64]byte{}
	m.Store("test", v)
	deleted := m.CompareAndDelete("test", v)
	assert.True(t, deleted)
	assert.Equal(t, getValueMapLen(m), 0)
}

func TestValueMapCompareAndDeleteNotEqual(t *testing.T) {
	m := new(ValueMap[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	v2 := &[64]byte{}
	deleted := m.CompareAndDelete("test", v2)
	assert.False(t, deleted)
	assert.Equal(t, getValueMapLen(m), 1)
	runtime.KeepAlive(v1)
	runtime.KeepAlive(v2)
}

func TestValueMapCompareAndDeleteNotFound(t *testing.T) {
	m := new(ValueMap[string, [64]byte])
	v := &[64]byte{}
	deleted := m.CompareAndDelete("test", v)
	assert.False(t, deleted)
	assert.Equal(t, getValueMapLen(m), 0)
	runtime.KeepAlive(v)
}

func BenchmarkValueMapCompareAndDelete(b *testing.B) {
	m := new(ValueMap[string, [64]byte])
	v := &[64]byte{}
	m.Store("test", v)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.CompareAndDelete("test", v)
		}
	})
	runtime.KeepAlive(v)
}

func TestValueMapCompareAndSwap(t *testing.T) {
	m := new(ValueMap[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	v2 := &[64]byte{}
	swapped := m.CompareAndSwap("test", v1, v2)
	assert.True(t, swapped)
	v3, ok := m.Load("test")
	assert.True(t, ok)
	assert.Equal(t, v3, v2)
	assert.Equal(t, getValueMapLen(m), 1)
	runtime.KeepAlive(v1)
	runtime.KeepAlive(v2)
}

func TestValueMapCompareAndSwapNotFound(t *testing.T) {
	m := new(ValueMap[string, [64]byte])
	v1 := &[64]byte{}
	v2 := &[64]byte{}
	swapped := m.CompareAndSwap("test", v2, v1)
	assert.False(t, swapped)
	_, ok := m.Load("test")
	assert.False(t, ok)
	assert.Equal(t, getValueMapLen(m), 0)
	runtime.KeepAlive(v1)
	runtime.KeepAlive(v2)
}

func TestValueMapCompareAndSwapNotEqual(t *testing.T) {
	m := new(ValueMap[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	v2 := &[64]byte{}
	swapped := m.CompareAndSwap("test", v2, v1)
	assert.False(t, swapped)
	v3, ok := m.Load("test")
	assert.True(t, ok)
	assert.Equal(t, v3, v1)
	assert.Equal(t, getValueMapLen(m), 1)
	runtime.KeepAlive(v1)
	runtime.KeepAlive(v2)
}

func TestValueMapCompareAndSwapSame(t *testing.T) {
	m := new(ValueMap[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	swapped := m.CompareAndSwap("test", v1, v1)
	assert.True(t, swapped)
	v2, ok := m.Load("test")
	assert.True(t, ok)
	assert.Equal(t, v2, v1)
	assert.Equal(t, getValueMapLen(m), 1)
	runtime.KeepAlive(v1)
}

func BenchmarkValueMapCompareAndSwapNotEqual(b *testing.B) {
	m := new(ValueMap[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	v2 := &[64]byte{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.CompareAndSwap("test", v2, v2)
		}
	})
	runtime.KeepAlive(v1)
}

func BenchmarkValueMapCompareAndSwapSame(b *testing.B) {
	m := new(ValueMap[string, [64]byte])
	v := &[64]byte{}
	m.Store("test", v)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.CompareAndSwap("test", v, v)
		}
	})
	runtime.KeepAlive(v)
}

func TestValueMapRange(t *testing.T) {
	m := new(ValueMap[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	found := false
	for k, v2 := range m.All() {
		assert.Equal(t, k, "test")
		assert.Equal(t, v2, v1)
		found = true
	}
	assert.True(t, found)
	assert.Equal(t, getValueMapLen(m), 1)
	runtime.KeepAlive(v1)
}

func TestValueMapRangeInterrupt(t *testing.T) {
	m := new(ValueMap[string, [64]byte])
	v := &[64]byte{}
	m.Store("test1", v)
	m.Store("test2", v)
	for range m.All() {
		break
	}
	runtime.KeepAlive(v)
}

func valueMapStoreDeadValue[K comparable, V any](m *ValueMap[K, V], key K) {
	v := new(V)
	runtime.SetFinalizer(v, func(*V) {}) // keeps the cleanup from running until a second GC
	m.Store(key, v)
	runtime.KeepAlive(v) // keep v alive through Store
}

func assertValueMapDeadEntry[K comparable, V any](tb testing.TB, m *ValueMap[K, V], key K) {
	tb.Helper()
	present, alive := ValueMapRawEntry(m, key)
	assert.True(tb, present)
	assert.False(tb, alive)
}

func assertValueMapNoEntry[K comparable, V any](tb testing.TB, m *ValueMap[K, V], key K) {
	tb.Helper()
	present, _ := ValueMapRawEntry(m, key)
	assert.False(tb, present)
}

func TestValueMapLoadEvictsDeadEntry(t *testing.T) {
	disableGC(t)
	m := new(ValueMap[string, [64]byte])
	valueMapStoreDeadValue(m, "test")
	runtime.GC() // v is unreachable: weak handle cleared (dead entry), cleanup kept until a second GC
	assertValueMapDeadEntry(t, m, "test")
	val, ok := m.Load("test")
	assert.False(t, ok)
	assert.Zero(t, val)
	assertValueMapNoEntry(t, m, "test")
}

func TestValueMapRangeEvictsDeadEntry(t *testing.T) {
	disableGC(t)
	m := new(ValueMap[string, [64]byte])
	valueMapStoreDeadValue(m, "test")
	runtime.GC() // v is unreachable: weak handle cleared (dead entry), cleanup kept until a second GC
	assertValueMapDeadEntry(t, m, "test")
	seen := false
	for range m.All() {
		seen = true
	}
	assert.False(t, seen)
	assertValueMapNoEntry(t, m, "test")
}

func TestValueMapCompareAndDeleteEvictsDeadEntry(t *testing.T) {
	disableGC(t)
	m := new(ValueMap[string, [64]byte])
	valueMapStoreDeadValue(m, "test")
	runtime.GC() // v is unreachable: dead entry, cleanup kept until a second GC
	assertValueMapDeadEntry(t, m, "test")
	assert.False(t, m.CompareAndDelete("test", &[64]byte{}))
	assertValueMapNoEntry(t, m, "test")
}

func TestValueMapCompareAndSwapEvictsDeadEntry(t *testing.T) {
	disableGC(t)
	m := new(ValueMap[string, [64]byte])
	valueMapStoreDeadValue(m, "test")
	runtime.GC() // v is unreachable: dead entry, cleanup kept until a second GC
	assertValueMapDeadEntry(t, m, "test")
	assert.False(t, m.CompareAndSwap("test", &[64]byte{}, &[64]byte{}))
	assertValueMapNoEntry(t, m, "test")
}

func BenchmarkValueMapRange(b *testing.B) {
	m := new(ValueMap[string, [64]byte])
	v := &[64]byte{}
	for i := range 10 {
		m.Store(fmt.Sprintf("test%d", i), v)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for range m.All() {
			}
		}
	})
	runtime.KeepAlive(v)
}

func getValueMapLen[K comparable, V any](m *ValueMap[K, V]) int {
	count := 0
	for range m.Range {
		count++
	}
	return count
}
