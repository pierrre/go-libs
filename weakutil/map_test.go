package weakutil_test

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/weakutil"
)

func ExampleMap() {
	m := new(Map[string, [64]byte])
	m.OnGCDelete = func(key string) {
		fmt.Println("GC delete:", key) // Shows that the key is deleted.
	}
	v := [64]byte{} // Must use a large value in order to trigger garbage collection reliably.
	m.Store("test", &v)
	runtime.GC()
	fmt.Println(m.Load("test")) // The pointer is still valid, because there is a keepalive below.
	runtime.KeepAlive(&v)
	runtime.GC()
	fmt.Println(m.Load("test")) // The pointer is not valid anymore.
	// Output:
	// &[0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0] true
	// GC delete: test
	// <nil> false
}

func TestMapLoad(t *testing.T) {
	m := new(Map[string, [64]byte]) // Must use a large value in order to trigger garbage collection reliably.
	v := [64]byte{}
	p1 := &v
	m.Store("test", p1)
	p2, ok := m.Load("test")
	assert.True(t, ok)
	assert.Equal(t, p1, p2)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(p1)
}

func TestMapLoadNil(t *testing.T) {
	m := new(Map[string, [64]byte])
	m.Store("test", nil)
	p, ok := m.Load("test")
	assert.True(t, ok)
	assert.Zero(t, p)
	assert.Equal(t, getMapLen(m), 1)
}

func TestMapLoadRemoved(t *testing.T) {
	m := new(Map[string, [64]byte])
	v := [64]byte{}
	m.Store("test", &v)
	runtime.GC()
	p, ok := m.Load("test")
	assert.False(t, ok)
	assert.Zero(t, p)
	assert.Equal(t, getMapLen(m), 0)
}

func TestMapStoreMultiple(t *testing.T) {
	for range 10 {
		m := new(Map[string, [64]byte])
		v := [64]byte{}
		p1 := &v
		wg := new(sync.WaitGroup)
		for range 100 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for range 100 {
					m.Store("test", p1)
				}
			}()
		}
		wg.Wait()
		p2, ok := m.Load("test")
		assert.True(t, ok)
		assert.Equal(t, p1, p2)
		assert.Equal(t, getMapLen(m), 1)
		runtime.KeepAlive(p1)
	}
}

func TestMapStoreMultipleDifferent(t *testing.T) {
	for range 10 {
		m := new(Map[string, [64]byte])
		wg := new(sync.WaitGroup)
		for range 100 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for range 100 {
					v := [64]byte{}
					m.Store("test", &v)
				}
			}()
		}
		wg.Wait()
		_, ok := m.Load("test")
		assert.True(t, ok)
		assert.LessOrEqual(t, getMapLen(m), 1)
	}
}

func TestMapDelete(t *testing.T) {
	m := new(Map[string, [64]byte])
	v := [64]byte{}
	p1 := &v
	m.Store("test", p1)
	m.Delete("test")
	p2, ok := m.Load("test")
	assert.False(t, ok)
	assert.Zero(t, p2)
	assert.Equal(t, getMapLen(m), 0)
	runtime.KeepAlive(p1)
}

func TestMapClear(t *testing.T) {
	m := new(Map[string, [64]byte])
	v := [64]byte{}
	p1 := &v
	m.Store("test", p1)
	m.Clear()
	p2, ok := m.Load("test")
	assert.False(t, ok)
	assert.Zero(t, p2)
	assert.Equal(t, getMapLen(m), 0)
	runtime.KeepAlive(p1)
}

func TestMapClean(t *testing.T) {
	m := new(Map[string, [64]byte])
	v := [64]byte{}
	m.Store("test", &v)
	runtime.GC()
	m.Clean()
	p, ok := m.Load("test")
	assert.False(t, ok)
	assert.Zero(t, p)
}

func TestMapOnGCDelete(t *testing.T) {
	m := new(Map[string, [64]byte])
	v := [64]byte{}
	m.Store("test", &v)
	runtime.GC()
	called := false
	m.OnGCDelete = func(key string) {
		called = true
		assert.Equal(t, key, "test")
	}
	_, ok := m.Load("test")
	assert.False(t, ok)
	assert.True(t, called)
}

func TestMapAutoClean(t *testing.T) {
	m := new(Map[string, [64]byte])
	for i := range 10 {
		m.Store(strconv.Itoa(i), nil)
	}
	autoDeleteCalled := false
	m.OnGCDelete = func(key string) {
		autoDeleteCalled = true
		assert.Equal(t, key, "test")
	}
	autoCleanCalled := false
	m.OnAutoClean = func() {
		autoCleanCalled = true
	}
	v := [64]byte{}
	m.Store("test", &v)
	runtime.GC()
	for range 1000 {
		m.Store("other", nil)
	}
	assert.True(t, autoDeleteCalled)
	assert.True(t, autoCleanCalled)
}

func TestMapAutoCleanAsync(t *testing.T) {
	m := new(Map[string, [64]byte])
	m.AutoCleanAsync = true
	called := false
	done := make(chan struct{})
	m.OnAutoClean = func() {
		if !called {
			called = true
			close(done)
		}
	}
	for range 1000 {
		m.Store("test", nil)
	}
	<-done
	assert.True(t, called)
}

func TestMapAutoCleanIntervalDisabled(t *testing.T) {
	m := new(Map[string, [64]byte])
	m.AutoCleanInterval = -1
	m.OnAutoClean = func() {
		t.Fatal("should not be called")
	}
	for range 1000 {
		m.Store("test", nil)
	}
}

func TestMapAutoCleanTryLockFail(t *testing.T) {
	m := new(Map[*[64]byte, [64]byte])
	m.AutoCleanInterval = 1
	wg := new(sync.WaitGroup)
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				v := [64]byte{}
				p := &v
				m.Store(p, p)
			}
		}()
	}
	wg.Wait()
	assert.Equal(t, getMapLen(m), 100*100)
}

func TestMapRange(t *testing.T) {
	m := new(Map[string, [64]byte])
	v := [64]byte{}
	p1 := &v
	m.Store("test", p1)
	found := false
	for k, p2 := range m.Range {
		assert.Equal(t, k, "test")
		assert.Equal(t, p2, p1)
		found = true
	}
	assert.True(t, found)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(p1)
}

func TestMapRangeRemoved(t *testing.T) {
	m := new(Map[string, [64]byte])
	v := [64]byte{}
	m.Store("test", &v)
	runtime.GC()
	for range m.Range {
		t.Fatal("should not range")
	}
	assert.Equal(t, getMapLen(m), 0)
}

func TestMapCompareAndDeleteTrue(t *testing.T) {
	m := new(Map[string, [64]byte])
	v := [64]byte{}
	m.Store("test", &v)
	deleted := m.CompareAndDelete("test", &v)
	assert.True(t, deleted)
	assert.Equal(t, getMapLen(m), 0)
}

func TestMapCompareAndDeleteFalse(t *testing.T) {
	m := new(Map[string, [64]byte])
	v1 := [64]byte{}
	p1 := &v1
	m.Store("test", &v1)
	v2 := [64]byte{}
	p2 := &v2
	deleted := m.CompareAndDelete("test", p2)
	assert.False(t, deleted)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(p1)
	runtime.KeepAlive(p2)
}

func TestMapCompareAndSwapTrue(t *testing.T) {
	m := new(Map[string, [64]byte])
	v1 := [64]byte{}
	p1 := &v1
	m.Store("test", p1)
	v2 := [64]byte{}
	p2 := &v2
	swapped := m.CompareAndSwap("test", p1, p2)
	assert.True(t, swapped)
	p3, ok := m.Load("test")
	assert.True(t, ok)
	assert.Equal(t, p3, p2)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(p1)
	runtime.KeepAlive(p2)
}

func TestMapCompareAndSwapFalse(t *testing.T) {
	m := new(Map[string, [64]byte])
	v1 := [64]byte{}
	p1 := &v1
	m.Store("test", p1)
	v2 := [64]byte{}
	p2 := &v2
	swapped := m.CompareAndSwap("test", p2, p1)
	assert.False(t, swapped)
	p3, ok := m.Load("test")
	assert.True(t, ok)
	assert.Equal(t, p3, p1)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(p1)
	runtime.KeepAlive(p2)
}

func TestMapLoadAndDeleteTrue(t *testing.T) {
	m := new(Map[string, [64]byte])
	v := [64]byte{}
	p1 := &v
	m.Store("test", p1)
	p2, loaded := m.LoadAndDelete("test")
	assert.True(t, loaded)
	assert.Equal(t, p1, p2)
	assert.Equal(t, getMapLen(m), 0)
}

func TestMapLoadAndDeleteFalse(t *testing.T) {
	m := new(Map[string, [64]byte])
	p, loaded := m.LoadAndDelete("test")
	assert.False(t, loaded)
	assert.Zero(t, p)
	assert.Equal(t, getMapLen(m), 0)
}

func TestMapLoadOrStoreTrue(t *testing.T) {
	m := new(Map[string, [64]byte])
	v1 := [64]byte{}
	p1 := &v1
	m.Store("test", p1)
	v2 := [64]byte{}
	p2 := &v2
	p3, loaded := m.LoadOrStore("test", p2)
	assert.True(t, loaded)
	assert.Equal(t, p3, p1)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(p1)
}

func TestMapLoadOrStoreFalse(t *testing.T) {
	m := new(Map[string, [64]byte])
	v := [64]byte{}
	p1 := &v
	p2, loaded := m.LoadOrStore("test", p1)
	assert.False(t, loaded)
	assert.Equal(t, p2, p1)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(p1)
}

func TestMapSwapTrue(t *testing.T) {
	m := new(Map[string, [64]byte])
	v1 := [64]byte{}
	p1 := &v1
	m.Store("test", p1)
	v2 := [64]byte{}
	p2 := &v2
	p3, loaded := m.Swap("test", p2)
	assert.True(t, loaded)
	assert.Equal(t, p3, p1)
	p4, ok := m.Load("test")
	assert.True(t, ok)
	assert.Equal(t, p4, p2)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(p1)
	runtime.KeepAlive(p2)
}

func TestMapSwapFalse(t *testing.T) {
	m := new(Map[string, [64]byte])
	v1 := [64]byte{}
	p1 := &v1
	p2, loaded := m.Swap("test", p1)
	assert.False(t, loaded)
	assert.Zero(t, p2)
	p3, ok := m.Load("test")
	assert.True(t, ok)
	assert.Equal(t, p3, p1)
	assert.Equal(t, getMapLen(m), 1)
	runtime.KeepAlive(p1)
}

var benchRes any

func BenchmarkMapStoreSame(b *testing.B) {
	m := new(Map[string, [64]byte])
	for i := range 1000 {
		m.Store(strconv.Itoa(i), nil)
	}
	v := [64]byte{}
	p := &v
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Store("test", p)
		}
	})
}

func BenchmarkMapStoreDifferent(b *testing.B) {
	m := new(Map[string, [64]byte])
	for i := range 1000 {
		m.Store(strconv.Itoa(i), nil)
	}
	v := [64]byte{}
	p := &v
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Store("test", nil)
			m.Store("test", p)
		}
	})
}

func BenchmarkMapStoreNil(b *testing.B) {
	m := new(Map[string, [64]byte])
	for i := range 1000 {
		m.Store(strconv.Itoa(i), nil)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Store("test", nil)
		}
	})
}

func BenchmarkMapLoad(b *testing.B) {
	m := new(Map[string, [64]byte])
	v := [64]byte{}
	p1 := &v
	m.Store("test", p1)
	var res *[64]byte
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			p2, ok := m.Load("test")
			assert.True(b, ok)
			assert.Equal(b, p2, p1)
			res = p2
		}
	})
	benchRes = res
}

func BenchmarkMapClean(b *testing.B) {
	m := new(Map[string, [64]byte])
	for i := range 1000 {
		m.Store(strconv.Itoa(i), nil)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Clean()
		}
	})
}

func getMapLen[K any, V any](m *Map[K, V]) int {
	count := 0
	for range m.Range {
		count++
	}
	return count
}
