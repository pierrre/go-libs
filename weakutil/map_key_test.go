package weakutil_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/weakutil"
)

func ExampleKeyMap() {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{} // Must use a large key in order to trigger garbage collection reliably.
	m.Store(k, "test")
	runtime.GC()
	fmt.Println(m.Load(k)) // The key is still valid, because there is a keepalive below.
	runtime.KeepAlive(k)
	runtime.GC()
	count := 0
	for range m.Range {
		count++
	}
	fmt.Println(count) // The key is not valid anymore.
	// Output:
	// test true
	// 0
}

func TestKeyMapStoreReplace(t *testing.T) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	v1 := "test1"
	m.Store(k, v1)
	v2 := "test2"
	m.Store(k, v2)
	v3, ok := m.Load(k)
	assert.True(t, ok)
	assert.Equal(t, v3, v2)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(k)
}

func TestKeyMapStoreNil(t *testing.T) {
	m := new(KeyMap[[64]byte, string])
	v1 := "test"
	m.Store(nil, v1)
	v2, ok := m.Load(nil)
	assert.True(t, ok)
	assert.Equal(t, v2, v1)
	assert.Equal(t, getMapLen(m), 1)
}

func TestKeyMapStoreNonComparable(t *testing.T) {
	m := new(KeyMap[[64]byte, []byte])
	k := &[64]byte{}
	v1 := []byte("test")
	m.Store(k, v1)
	m.Store(k, v1) // Must not panic.
	_, ok := m.Load(k)
	assert.True(t, ok)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(k)
}

func BenchmarkKeyMapStoreSame(b *testing.B) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	v := "test"
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Store(k, v)
		}
	})
}

func BenchmarkKeyMapStoreDifferent(b *testing.B) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		vs := [2]string{"test1", "test2"}
		for i := 0; pb.Next(); i++ {
			m.Store(k, vs[i%2])
		}
	})
}

func BenchmarkKeyMapStoreNewRandomKey(b *testing.B) {
	m := new(KeyMap[[64]byte, string])
	v := "test"
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			k := &[64]byte{}
			m.Store(k, v)
		}
	})
}

func TestKeyMapLoadNotFound(t *testing.T) {
	m := new(KeyMap[[64]byte, string])
	v, ok := m.Load(nil)
	assert.False(t, ok)
	assert.Zero(t, v)
	assert.Equal(t, getMapLen(m), 0)
}

func BenchmarkKeyMapLoad(b *testing.B) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	v := "test"
	m.Store(k, v)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = m.Load(k)
		}
	})
}

func BenchmarkKeyMapLoadNil(b *testing.B) {
	m := new(KeyMap[[64]byte, string])
	v := "test"
	m.Store(nil, v)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = m.Load(nil)
		}
	})
}

func BenchmarkKeyMapLoadNotFound(b *testing.B) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = m.Load(k)
		}
	})
}

func TestKeyMapDelete(t *testing.T) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	v1 := "test"
	m.Store(k, v1)
	m.Delete(k)
	v2, ok := m.Load(k)
	assert.False(t, ok)
	assert.Zero(t, v2)
	assert.Equal(t, getMapLen(m), 0)
	runtime.KeepAlive(k)
}

func TestKeyMapDeleteNotFound(t *testing.T) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	m.Delete(k)
	assert.Equal(t, getMapLen(m), 0)
	runtime.KeepAlive(k)
}

func BenchmarkKeyMapDelete(b *testing.B) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Delete(k)
		}
	})
}

func TestKeyMapClear(t *testing.T) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	v1 := "test"
	m.Store(k, v1)
	m.Clear()
	v2, ok := m.Load(k)
	assert.False(t, ok)
	assert.Zero(t, v2)
	assert.Equal(t, getMapLen(m), 0)
	runtime.KeepAlive(v1)
}

func TestKeyMapClearEmpty(t *testing.T) {
	m := new(KeyMap[[64]byte, string])
	m.Clear()
	assert.Equal(t, getMapLen(m), 0)
}

func TestKeyMapClearNonComparable(t *testing.T) {
	m := new(KeyMap[[64]byte, []byte])
	k := &[64]byte{}
	v := []byte("test")
	m.Store(k, v)
	m.Clear()
	v2, ok := m.Load(k)
	assert.False(t, ok)
	assert.SliceEmpty(t, v2)
	assert.Equal(t, getMapLen(m), 0)
	runtime.KeepAlive(k)
}

func BenchmarkKeyMapClear(b *testing.B) {
	m := new(KeyMap[[64]byte, string])
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Clear()
		}
	})
}

func TestKeyMapSwap(t *testing.T) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	v1 := "test1"
	m.Store(k, v1)
	v2 := "test2"
	v3, loaded := m.Swap(k, v2)
	assert.True(t, loaded)
	assert.Equal(t, v3, v1)
	v4, ok := m.Load(k)
	assert.True(t, ok)
	assert.Equal(t, v4, v2)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(k)
}

func TestKeyMapSwapSame(t *testing.T) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	v1 := "test"
	m.Store(k, v1)
	v2, loaded := m.Swap(k, v1)
	assert.True(t, loaded)
	assert.Equal(t, v2, v1)
	v3, ok := m.Load(k)
	assert.True(t, ok)
	assert.Equal(t, v3, v1)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(k)
}

func TestKeyMapSwapNotFound(t *testing.T) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	v1 := "test"
	v2, loaded := m.Swap(k, v1)
	assert.False(t, loaded)
	assert.Zero(t, v2)
	v3, ok := m.Load(k)
	assert.True(t, ok)
	assert.Equal(t, v3, v1)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(k)
}

func TestKeyMapSwapNonComparable(t *testing.T) {
	m := new(KeyMap[[64]byte, []byte])
	k := &[64]byte{}
	v1 := []byte("test")
	m.Store(k, v1)
	v2, loaded := m.Swap(k, v1) // Must not panic.
	assert.True(t, loaded)
	assert.DeepEqual(t, v2, v1)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(k)
}

func TestKeyMapLoadAndDelete(t *testing.T) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	v1 := "test"
	m.Store(k, v1)
	v2, loaded := m.LoadAndDelete(k)
	assert.True(t, loaded)
	assert.Equal(t, v2, v1)
	assert.Equal(t, getMapLen(m), 0)
	runtime.KeepAlive(k)
}

func TestKeyMapLoadAndDeleteNotFound(t *testing.T) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	v, loaded := m.LoadAndDelete(k)
	assert.False(t, loaded)
	assert.Zero(t, v)
	assert.Equal(t, getMapLen(m), 0)
	runtime.KeepAlive(k)
}

func BenchmarkKeyMapLoadAndDelete(b *testing.B) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.LoadAndDelete(k)
		}
	})
	runtime.KeepAlive(k)
}

func TestKeyMapLoadOrStore(t *testing.T) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	v1 := "test1"
	m.Store(k, v1)
	v2 := "test2"
	v3, loaded := m.LoadOrStore(k, v2)
	assert.True(t, loaded)
	assert.Equal(t, v3, v1)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(k)
}

func TestKeyMapLoadOrStoreNotFound(t *testing.T) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	v1 := "test"
	v2, loaded := m.LoadOrStore(k, v1)
	assert.False(t, loaded)
	assert.Equal(t, v2, v1)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(k)
}

func BenchmarkKeyMapLoadOrStore(b *testing.B) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	v := "test"
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.LoadOrStore(k, v)
		}
	})
	runtime.KeepAlive(k)
}

func TestKeyMapCompareAndDelete(t *testing.T) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	v := "test"
	m.Store(k, v)
	deleted := m.CompareAndDelete(k, v)
	assert.True(t, deleted)
	assert.Equal(t, getMapLen(m), 0)
	runtime.KeepAlive(k)
}

func TestKeyMapCompareAndDeleteNotEqual(t *testing.T) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	v1 := "test1"
	m.Store(k, v1)
	v2 := "test2"
	deleted := m.CompareAndDelete(k, v2)
	assert.False(t, deleted)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(k)
}

func TestKeyMapCompareAndDeleteNotFound(t *testing.T) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	v := "test"
	deleted := m.CompareAndDelete(k, v)
	assert.False(t, deleted)
	assert.Equal(t, getMapLen(m), 0)
	runtime.KeepAlive(k)
}

func BenchmarkKeyMapCompareAndDelete(b *testing.B) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	v := "test"
	m.Store(k, v)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.CompareAndDelete(k, v)
		}
	})
	runtime.KeepAlive(k)
}

func TestKeyMapCompareAndSwap(t *testing.T) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	v1 := "test1"
	m.Store(k, v1)
	v2 := "test2"
	swapped := m.CompareAndSwap(k, v1, v2)
	assert.True(t, swapped)
	v3, ok := m.Load(k)
	assert.True(t, ok)
	assert.Equal(t, v3, v2)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(k)
}

func TestKeyMapCompareAndSwapNotFound(t *testing.T) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	v1 := "test1"
	v2 := "test2"
	swapped := m.CompareAndSwap(k, v2, v1)
	assert.False(t, swapped)
	_, ok := m.Load(k)
	assert.False(t, ok)
	assert.Equal(t, getMapLen(m), 0)
	runtime.KeepAlive(k)
}

func TestKeyMapCompareAndSwapNotEqual(t *testing.T) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	v1 := "test1"
	m.Store(k, v1)
	v2 := "test2"
	swapped := m.CompareAndSwap(k, v2, v1)
	assert.False(t, swapped)
	v3, ok := m.Load(k)
	assert.True(t, ok)
	assert.Equal(t, v3, v1)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(k)
}

func TestKeyMapCompareAndSwapEqual(t *testing.T) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	v1 := "test"
	m.Store(k, v1)
	swapped := m.CompareAndSwap(k, v1, v1)
	assert.True(t, swapped)
	v2, ok := m.Load(k)
	assert.True(t, ok)
	assert.Equal(t, v2, v1)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(k)
}

func BenchmarkKeyMapCompareAndSwapNotEqual(b *testing.B) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	v1 := "test1"
	m.Store(k, v1)
	v2 := "test2"
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.CompareAndSwap(k, v2, v2)
		}
	})
	runtime.KeepAlive(k)
}

func BenchmarkKeyMapCompareAndSwapEqual(b *testing.B) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	v := "test"
	m.Store(k, v)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.CompareAndSwap(k, v, v)
		}
	})
	runtime.KeepAlive(k)
}

func TestKeyMapRange(t *testing.T) {
	m := new(KeyMap[[64]byte, string])
	k := &[64]byte{}
	v1 := "test"
	m.Store(k, v1)
	found := false
	for k2, v2 := range m.All() {
		assert.Equal(t, k2, k)
		assert.Equal(t, v2, v1)
		found = true
	}
	assert.True(t, found)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(k)
}

func TestKeyMapRangeNil(t *testing.T) {
	m := new(KeyMap[[64]byte, string])
	m.Store(nil, "test")
	found := false
	for k, v := range m.All() {
		assert.Zero(t, k)
		assert.Equal(t, v, "test")
		found = true
	}
	assert.True(t, found)
}

func TestKeyMapRangeInterrupt(t *testing.T) {
	m := new(KeyMap[[64]byte, string])
	k1 := &[64]byte{}
	k2 := &[64]byte{}
	m.Store(k1, "test1")
	m.Store(k2, "test2")
	for range m.All() {
		break
	}
	runtime.KeepAlive(k1)
	runtime.KeepAlive(k2)
}

func TestKeyMapRangeEvictsDeadKey(t *testing.T) {
	disableGC(t)
	m := new(KeyMap[[64]byte, int]) // must use a large key in order to trigger garbage collection reliably
	k := &[64]byte{}
	runtime.SetFinalizer(k, func(*[64]byte) {}) // keeps the cleanup from running until a second GC
	m.Store(k, 123)
	runtime.KeepAlive(k) // keep k alive through Store
	runtime.GC()         // k is unreachable: weak handle cleared (dead entry), cleanup kept until a second GC
	total, dead := KeyMapRawEntries(m)
	assert.Equal(t, total, 1)
	assert.Equal(t, dead, 1)
	seen := false
	for range m.All() {
		seen = true
	}
	assert.False(t, seen)
	total, _ = KeyMapRawEntries(m)
	assert.Equal(t, total, 0)
}

func TestKeyMapRangeEvictsDeadKeyNonComparable(t *testing.T) {
	disableGC(t)
	m := new(KeyMap[[64]byte, []byte]) // must use a large key in order to trigger garbage collection reliably
	k := &[64]byte{}
	runtime.SetFinalizer(k, func(*[64]byte) {}) // keeps the cleanup from running until a second GC
	m.Store(k, []byte("test"))
	runtime.KeepAlive(k) // keep k alive through Store
	runtime.GC()         // k is unreachable: weak handle cleared (dead entry), cleanup kept until a second GC
	total, dead := KeyMapRawEntries(m)
	assert.Equal(t, total, 1)
	assert.Equal(t, dead, 1)
	seen := false
	for range m.All() {
		seen = true
	}
	assert.False(t, seen)
	total, _ = KeyMapRawEntries(m)
	assert.Equal(t, total, 0)
}

func BenchmarkKeyMapRange(b *testing.B) {
	m := new(KeyMap[[64]byte, string])
	var ks [10]*[64]byte
	for i := range 10 {
		k := &[64]byte{}
		ks[i] = k
		m.Store(k, fmt.Sprintf("test%d", i))
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for range m.All() {
			}
		}
	})
	runtime.KeepAlive(ks)
}
