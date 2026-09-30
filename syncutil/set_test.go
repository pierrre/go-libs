package syncutil_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/syncutil"
)

func TestSetAdd(t *testing.T) {
	var s Set[string]
	assert.True(t, s.Add("key"))
	assert.False(t, s.Add("key"))
	assert.True(t, s.Contains("key"))
}

func TestSetContains(t *testing.T) {
	var s Set[string]
	assert.False(t, s.Contains("key"))
	s.Add("key")
	assert.True(t, s.Contains("key"))
	s.Delete("key")
	assert.False(t, s.Contains("key"))
}

func TestSetDelete(t *testing.T) {
	var s Set[string]
	assert.False(t, s.Delete("key"))
	s.Add("key")
	assert.True(t, s.Delete("key"))
	assert.False(t, s.Delete("key"))
}

func TestSetAll(t *testing.T) {
	var s Set[string]
	s.Add("a")
	s.Add("b")
	s.Add("c")
	var keys []string
	for key := range s.All() {
		keys = append(keys, key)
	}
	assert.SliceElementsMatch(t, keys, []string{"a", "b", "c"})
	n := 0
	for range s.All() {
		n++
		break
	}
	assert.Equal(t, n, 1)
}

func TestSetClear(t *testing.T) {
	var s Set[string]
	s.Add("a")
	s.Add("b")
	s.Clear()
	assert.False(t, s.Contains("a"))
	assert.False(t, s.Contains("b"))
}

func TestSetConcurrentAdd(t *testing.T) {
	var s Set[string]
	var added atomic.Int64
	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() {
			if s.Add("key") {
				added.Add(1)
			}
		})
	}
	wg.Wait()
	assert.Equal(t, added.Load(), 1)
	assert.True(t, s.Contains("key"))
}

func TestSetConcurrentDelete(t *testing.T) {
	var s Set[string]
	s.Add("key")
	var deleted atomic.Int64
	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() {
			if s.Delete("key") {
				deleted.Add(1)
			}
		})
	}
	wg.Wait()
	assert.Equal(t, deleted.Load(), 1)
	assert.False(t, s.Contains("key"))
}

func BenchmarkSetAdd(b *testing.B) {
	var s Set[string]
	for b.Loop() {
		s.Add("key")
	}
}

func BenchmarkSetAddParallel(b *testing.B) {
	var s Set[string]
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			s.Add("key")
		}
	})
}

func BenchmarkSetContains(b *testing.B) {
	var s Set[string]
	s.Add("key")
	for b.Loop() {
		s.Contains("key")
	}
}

func BenchmarkSetContainsParallel(b *testing.B) {
	var s Set[string]
	s.Add("key")
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			s.Contains("key")
		}
	})
}

func BenchmarkSetDelete(b *testing.B) {
	var s Set[string]
	s.Add("key")
	for b.Loop() {
		s.Delete("key")
	}
}

func BenchmarkSetDeleteParallel(b *testing.B) {
	var s Set[string]
	s.Add("key")
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			s.Delete("key")
		}
	})
}
