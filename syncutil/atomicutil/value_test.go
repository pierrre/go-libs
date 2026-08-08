package atomicutil_test

import (
	"sync/atomic"
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/syncutil/atomicutil"
)

func TestValue(t *testing.T) {
	var v Value[int]
	assert.Zero(t, v.Load())
	v.Store(1)
	assert.Equal(t, v.Load(), 1)
	swapped := v.CompareAndSwap(1, 2)
	assert.True(t, swapped)
	assert.Equal(t, v.Load(), 2)
	swapped = v.CompareAndSwap(1, 3)
	assert.False(t, swapped)
	assert.Equal(t, v.Load(), 2)
	old := v.Swap(3)
	assert.Equal(t, old, 2)
	assert.Equal(t, v.Load(), 3)
}

func TestValueCompareAndSwapPanicNonComparable(t *testing.T) {
	var v Value[map[string]int]
	v.Store(map[string]int{"a": 1})
	assert.Panics(t, func() {
		v.CompareAndSwap(map[string]int{"a": 1}, map[string]int{"a": 2})
	})
}

func BenchmarkValueStore(b *testing.B) {
	var v Value[int]
	for b.Loop() {
		v.Store(1)
	}
}

func BenchmarkValueStoreParallel(b *testing.B) {
	var v Value[int]
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			v.Store(1)
		}
	})
}

func BenchmarkSyncAtomicValueStore(b *testing.B) {
	var v atomic.Value
	for b.Loop() {
		v.Store(1)
	}
}

func BenchmarkSyncAtomicValueStoreParallel(b *testing.B) {
	var v atomic.Value
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			v.Store(1)
		}
	})
}

func BenchmarkValueLoad(b *testing.B) {
	var v Value[int]
	v.Store(1)
	for b.Loop() {
		v.Load()
	}
}

func BenchmarkValueLoadParallel(b *testing.B) {
	var v Value[int]
	v.Store(1)
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			v.Load()
		}
	})
}

func BenchmarkSyncAtomicValueLoad(b *testing.B) {
	var v atomic.Value
	v.Store(1)
	for b.Loop() {
		v.Load()
	}
}

func BenchmarkSyncAtomicValueLoadParallel(b *testing.B) {
	var v atomic.Value
	v.Store(1)
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			v.Load()
		}
	})
}
