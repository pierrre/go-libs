package weakutil

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/pierrre/assert"
)

func ExampleSet() {
	s := new(Set[[64]byte])
	k := &[64]byte{} // Must use a large key in order to trigger garbage collection reliably.
	s.Add(k)
	runtime.GC()
	fmt.Println(s.Contains(k)) // The key is still valid, because there is a keepalive below.
	runtime.KeepAlive(k)
	runtime.GC()
	count := 0
	for range s.All() {
		count++
	}
	fmt.Println(count) // The key is not valid anymore.
	// Output:
	// true
	// 0
}

func TestSetAdd(t *testing.T) {
	s := new(Set[[64]byte])
	k := &[64]byte{}
	assert.True(t, s.Add(k))
	assert.False(t, s.Add(k))
	assert.True(t, s.Contains(k))
	runtime.KeepAlive(k)
}

func TestSetContains(t *testing.T) {
	s := new(Set[[64]byte])
	k := &[64]byte{}
	assert.False(t, s.Contains(k))
	assert.True(t, s.Add(k))
	assert.True(t, s.Contains(k))
	assert.True(t, s.Delete(k))
	assert.False(t, s.Contains(k))
	runtime.KeepAlive(k)
}

func TestSetDelete(t *testing.T) {
	s := new(Set[[64]byte])
	k := &[64]byte{}
	assert.False(t, s.Delete(k))
	assert.True(t, s.Add(k))
	assert.True(t, s.Delete(k))
	assert.False(t, s.Delete(k))
	runtime.KeepAlive(k)
}

func TestSetAll(t *testing.T) {
	s := new(Set[[64]byte])
	var ks [3]*[64]byte
	for i := range ks {
		k := &[64]byte{}
		ks[i] = k
		s.Add(k)
	}
	var keys []*[64]byte
	for key := range s.All() {
		keys = append(keys, key)
	}
	assert.SliceElementsMatch(t, keys, ks[:])
	n := 0
	for range s.All() {
		n++
		break
	}
	assert.Equal(t, n, 1)
	runtime.KeepAlive(ks)
}

func TestSetClear(t *testing.T) {
	s := new(Set[[64]byte])
	k1 := &[64]byte{}
	k2 := &[64]byte{}
	s.Add(k1)
	s.Add(k2)
	s.Clear()
	assert.False(t, s.Contains(k1))
	assert.False(t, s.Contains(k2))
	runtime.KeepAlive(k1)
	runtime.KeepAlive(k2)
}

func TestSetNilKey(t *testing.T) {
	s := new(Set[[64]byte])
	assert.True(t, s.Add(nil))
	assert.False(t, s.Add(nil))
	assert.True(t, s.Contains(nil))
	found := false
	for key := range s.All() {
		assert.Zero(t, key)
		found = true
	}
	assert.True(t, found)
}

func TestSetConcurrentAdd(t *testing.T) {
	s := new(Set[[64]byte])
	k := &[64]byte{}
	var added atomic.Int64
	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() {
			if s.Add(k) {
				added.Add(1)
			}
		})
	}
	wg.Wait()
	assert.Equal(t, added.Load(), 1)
	assert.True(t, s.Contains(k))
	runtime.KeepAlive(k)
}

func TestSetConcurrentDelete(t *testing.T) {
	s := new(Set[[64]byte])
	k := &[64]byte{}
	s.Add(k)
	var deleted atomic.Int64
	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() {
			if s.Delete(k) {
				deleted.Add(1)
			}
		})
	}
	wg.Wait()
	assert.Equal(t, deleted.Load(), 1)
	assert.False(t, s.Contains(k))
	runtime.KeepAlive(k)
}

func TestSetRangeEvictsDeadKey(t *testing.T) {
	disableGC(t)
	s := new(Set[[64]byte]) // must use a large key in order to trigger garbage collection reliably
	k := &[64]byte{}
	runtime.SetFinalizer(k, func(*[64]byte) {}) // keeps the cleanup from running until a second GC
	s.Add(k)
	runtime.KeepAlive(k) // keep k alive through Add
	runtime.GC()         // k is unreachable: weak handle cleared (dead entry), cleanup kept until a second GC
	total, dead := keyMapRawEntries(&s.m)
	assert.Equal(t, total, 1)
	assert.Equal(t, dead, 1)
	seen := false
	for range s.All() {
		seen = true
	}
	assert.False(t, seen)
	total, _ = keyMapRawEntries(&s.m)
	assert.Equal(t, total, 0)
}

func BenchmarkSetAdd(b *testing.B) {
	s := new(Set[[64]byte])
	k := &[64]byte{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s.Add(k)
		}
	})
	runtime.KeepAlive(k)
}

func BenchmarkSetContains(b *testing.B) {
	s := new(Set[[64]byte])
	k := &[64]byte{}
	s.Add(k)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s.Contains(k)
		}
	})
	runtime.KeepAlive(k)
}

func BenchmarkSetDelete(b *testing.B) {
	s := new(Set[[64]byte])
	k := &[64]byte{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s.Delete(k)
		}
	})
	runtime.KeepAlive(k)
}
