package weakutil_test

import (
	"fmt"
	"math/rand/v2"
	"runtime"
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/weakutil"
)

func ExampleMap() {
	m := new(Map[string, [64]byte])
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

func TestMapStoreLoad(t *testing.T) {
	m := new(Map[string, [64]byte]) // Must use a large value in order to trigger garbage collection reliably.
	v1 := &[64]byte{}
	m.Store("test", v1)
	m.Store("test", v1)
	v2, ok := m.Load("test")
	assert.True(t, ok)
	assert.Equal(t, v1, v2)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(v1)
}

func TestMapStoreReplace(t *testing.T) {
	m := new(Map[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	v2 := &[64]byte{}
	m.Store("test", v2)
	v3, ok := m.Load("test")
	assert.True(t, ok)
	assert.Equal(t, v2, v3)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(v1)
	runtime.KeepAlive(v2)
}

func BenchmarkMapStoreSame(b *testing.B) {
	m := new(Map[string, [64]byte])
	v := &[64]byte{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Store("test", v)
		}
	})
}

func BenchmarkMapStoreDifferent(b *testing.B) {
	m := new(Map[string, [64]byte])
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var vs [2][64]byte
		for i := 0; pb.Next(); i++ {
			v := &vs[i%2]
			m.Store("test", v)
		}
	})
}

func BenchmarkMapStoreNewRandomKey(b *testing.B) {
	m := new(Map[int64, [64]byte])
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Store(rand.Int64(), &[64]byte{}) //nolint:gosec // This rand package is OK, it's a test.
		}
	})
}

func TestMapLoadNil(t *testing.T) {
	m := new(Map[string, [64]byte])
	m.Store("test", nil)
	v, ok := m.Load("test")
	assert.True(t, ok)
	assert.Zero(t, v)
	assert.Equal(t, getMapLen(m), 1)
}

func TestMapLoadNotFound(t *testing.T) {
	m := new(Map[string, [64]byte])
	v, ok := m.Load("test")
	assert.False(t, ok)
	assert.Zero(t, v)
	assert.Equal(t, getMapLen(m), 0)
}

func TestMapLoadRemovedGC(t *testing.T) {
	m := new(Map[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	runtime.GC()
	v2, ok := m.Load("test")
	assert.False(t, ok)
	assert.Zero(t, v2)
	assert.Equal(t, getMapLen(m), 0)
}

func BenchmarkMapLoad(b *testing.B) {
	m := new(Map[string, [64]byte])
	v := &[64]byte{}
	m.Store("test", v)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Load("test")
		}
	})
}

func BenchmarkMapLoadNil(b *testing.B) {
	m := new(Map[string, [64]byte])
	m.Store("test", nil)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Load("test")
		}
	})
}

func BenchmarkMapLoadNotFound(b *testing.B) {
	m := new(Map[string, [64]byte])
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Load("test")
		}
	})
}

func TestMapDelete(t *testing.T) {
	m := new(Map[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	m.Delete("test")
	v2, ok := m.Load("test")
	assert.False(t, ok)
	assert.Zero(t, v2)
	assert.Equal(t, getMapLen(m), 0)
	runtime.KeepAlive(v1)
}

func TestMapDeleteNotFound(t *testing.T) {
	m := new(Map[string, [64]byte])
	m.Delete("test")
	assert.Equal(t, getMapLen(m), 0)
}

func BenchmarkMapDelete(b *testing.B) {
	m := new(Map[string, [64]byte])
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Delete("test")
		}
	})
}

func TestMapClear(t *testing.T) {
	m := new(Map[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	m.Clear()
	v2, ok := m.Load("test")
	assert.False(t, ok)
	assert.Zero(t, v2)
	assert.Equal(t, getMapLen(m), 0)
	runtime.KeepAlive(v1)
}

func TestMapClearEmpty(t *testing.T) {
	m := new(Map[string, [64]byte])
	m.Clear()
	assert.Equal(t, getMapLen(m), 0)
}

func BenchmarkMapClear(b *testing.B) {
	m := new(Map[string, [64]byte])
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Clear()
		}
	})
}

func TestMapSwap(t *testing.T) {
	m := new(Map[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	v2 := &[64]byte{}
	v3, loaded := m.Swap("test", v2)
	assert.True(t, loaded)
	assert.Equal(t, v3, v1)
	v4, ok := m.Load("test")
	assert.True(t, ok)
	assert.Equal(t, v4, v2)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(v1)
	runtime.KeepAlive(v2)
}

func TestMapSwapSame(t *testing.T) {
	m := new(Map[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	v2, loaded := m.Swap("test", v1)
	assert.True(t, loaded)
	assert.Equal(t, v2, v1)
	v3, ok := m.Load("test")
	assert.True(t, ok)
	assert.Equal(t, v3, v1)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(v1)
}

func TestMapSwapNotFound(t *testing.T) {
	m := new(Map[string, [64]byte])
	v1 := &[64]byte{}
	v2, loaded := m.Swap("test", v1)
	assert.False(t, loaded)
	assert.Zero(t, v2)
	v3, ok := m.Load("test")
	assert.True(t, ok)
	assert.Equal(t, v3, v1)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(v1)
}

func BenchmarkMapSwapSame(b *testing.B) {
	m := new(Map[string, [64]byte])
	v := &[64]byte{}
	m.Store("test", v)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Swap("test", v)
		}
	})
}

func BenchmarkMapSwapDifferent(b *testing.B) {
	m := new(Map[string, [64]byte])
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var vs [2][64]byte
		for i := 0; pb.Next(); i++ {
			v := &vs[i%2]
			m.Swap("test", v)
		}
	})
}

func TestMapLoadAndDelete(t *testing.T) {
	m := new(Map[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	v2, loaded := m.LoadAndDelete("test")
	assert.True(t, loaded)
	assert.Equal(t, v1, v2)
	assert.Equal(t, getMapLen(m), 0)
}

func TestMapLoadAndDeleteNotFound(t *testing.T) {
	m := new(Map[string, [64]byte])
	v, loaded := m.LoadAndDelete("test")
	assert.False(t, loaded)
	assert.Zero(t, v)
	assert.Equal(t, getMapLen(m), 0)
}

func BenchmarkMapLoadAndDelete(b *testing.B) {
	m := new(Map[string, [64]byte])
	v := &[64]byte{}
	m.Store("test", v)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.LoadAndDelete("test")
		}
	})
	runtime.KeepAlive(v)
}

func TestMapLoadOrStore(t *testing.T) {
	m := new(Map[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	v2 := &[64]byte{}
	v3, loaded := m.LoadOrStore("test", v2)
	assert.True(t, loaded)
	assert.Equal(t, v3, v1)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(v1)
}

func TestMapLoadOrStoreNotFound(t *testing.T) {
	m := new(Map[string, [64]byte])
	v1 := &[64]byte{}
	v2, loaded := m.LoadOrStore("test", v1)
	assert.False(t, loaded)
	assert.Equal(t, v2, v1)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(v1)
}

func BenchmarkMapLoadOrStore(b *testing.B) {
	m := new(Map[string, [64]byte])
	v := &[64]byte{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.LoadOrStore("test", v)
		}
	})
	runtime.KeepAlive(v)
}

func TestMapCompareAndDelete(t *testing.T) {
	m := new(Map[string, [64]byte])
	v := &[64]byte{}
	m.Store("test", v)
	deleted := m.CompareAndDelete("test", v)
	assert.True(t, deleted)
	assert.Equal(t, getMapLen(m), 0)
}

func TestMapCompareAndDeleteNotEqual(t *testing.T) {
	m := new(Map[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	v2 := &[64]byte{}
	deleted := m.CompareAndDelete("test", v2)
	assert.False(t, deleted)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(v1)
	runtime.KeepAlive(v2)
}

func TestMapCompareAndDeleteNotFound(t *testing.T) {
	m := new(Map[string, [64]byte])
	v := &[64]byte{}
	deleted := m.CompareAndDelete("test", v)
	assert.False(t, deleted)
	assert.Equal(t, getMapLen(m), 0)
	runtime.KeepAlive(v)
}

func BenchmarkMapCompareAndDelete(b *testing.B) {
	m := new(Map[string, [64]byte])
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

func TestMapCompareAndSwap(t *testing.T) {
	m := new(Map[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	v2 := &[64]byte{}
	swapped := m.CompareAndSwap("test", v1, v2)
	assert.True(t, swapped)
	v3, ok := m.Load("test")
	assert.True(t, ok)
	assert.Equal(t, v3, v2)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(v1)
	runtime.KeepAlive(v2)
}

func TestMapCompareAndSwapNotFound(t *testing.T) {
	m := new(Map[string, [64]byte])
	v1 := &[64]byte{}
	v2 := &[64]byte{}
	swapped := m.CompareAndSwap("test", v2, v1)
	assert.False(t, swapped)
	_, ok := m.Load("test")
	assert.False(t, ok)
	assert.Equal(t, getMapLen(m), 0)
	runtime.KeepAlive(v1)
	runtime.KeepAlive(v2)
}

func TestMapCompareAndSwapNotEqual(t *testing.T) {
	m := new(Map[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	v2 := &[64]byte{}
	swapped := m.CompareAndSwap("test", v2, v1)
	assert.False(t, swapped)
	v3, ok := m.Load("test")
	assert.True(t, ok)
	assert.Equal(t, v3, v1)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(v1)
	runtime.KeepAlive(v2)
}

func TestMapCompareAndSwapSame(t *testing.T) {
	m := new(Map[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	swapped := m.CompareAndSwap("test", v1, v1)
	assert.True(t, swapped)
	v2, ok := m.Load("test")
	assert.True(t, ok)
	assert.Equal(t, v2, v1)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(v1)
}

func BenchmarkMapCompareAndSwapNotEqual(b *testing.B) {
	m := new(Map[string, [64]byte])
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

func BenchmarkMapCompareAndSwapSame(b *testing.B) {
	m := new(Map[string, [64]byte])
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

func TestMapRange(t *testing.T) {
	m := new(Map[string, [64]byte])
	v1 := &[64]byte{}
	m.Store("test", v1)
	found := false
	for k, v2 := range m.All() {
		assert.Equal(t, k, "test")
		assert.Equal(t, v2, v1)
		found = true
	}
	assert.True(t, found)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(v1)
}

func TestMapRangeInterrupt(t *testing.T) {
	m := new(Map[string, [64]byte])
	v := &[64]byte{}
	m.Store("test1", v)
	m.Store("test2", v)
	for range m.All() {
		break
	}
	runtime.KeepAlive(v)
}

func BenchmarkMapRange(b *testing.B) {
	m := new(Map[string, [64]byte])
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

func getMapLen[K comparable, V any](m *Map[K, V]) int {
	count := 0
	for range m.Range {
		count++
	}
	return count
}
