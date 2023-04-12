package goroutine

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/pierrre/assert"
	"github.com/pierrre/go-libs/internal/golibstest"
)

func init() {
	golibstest.Configure()
}

func TestGo(t *testing.T) {
	var called int64
	done := make(chan struct{})
	Go(func() {
		atomic.AddInt64(&called, 1)
		close(done)
	})
	<-done
	assert.Equal(t, called, 1)
}

func TestGoAllocs(t *testing.T) {
	done := make(chan struct{})
	assert.AllocsPerRun(t, 100, func() {
		Go(func() {
			done <- struct{}{}
		})
		<-done
	}, 2)
}

func TestGoWait(t *testing.T) {
	var called int64
	wait := GoWait(func() {
		atomic.AddInt64(&called, 1)
	})
	wait()
	assert.Equal(t, called, 1)
}

func TestGoWaitAllocs(t *testing.T) {
	assert.AllocsPerRun(t, 100, func() {
		wait := GoWait(func() {})
		wait()
	}, 4)
}

func TestWaitGroup(t *testing.T) {
	wg := new(sync.WaitGroup)
	var called int64
	WaitGroup(wg, func() {
		atomic.AddInt64(&called, 1)
	})
	wg.Wait()
	assert.Equal(t, called, 1)
}

func TestWaitGroupAllocs(t *testing.T) {
	wg := new(sync.WaitGroup)
	assert.AllocsPerRun(t, 100, func() {
		WaitGroup(wg, func() {})
		wg.Wait()
	}, 2)
}

func TestN(t *testing.T) {
	count := 10
	mu := new(sync.Mutex)
	is := make(map[int]struct{})
	N(count, func(i int) {
		mu.Lock()
		defer mu.Unlock()
		is[i] = struct{}{}
	})
	isExpected := make(map[int]struct{})
	for i := 0; i < count; i++ {
		isExpected[i] = struct{}{}
	}
	assert.MapEqual(t, is, isExpected)
}

func TestSlice(t *testing.T) {
	s := []string{"a", "b", "c"}
	mu := new(sync.Mutex)
	res := make([]string, len(s))
	Slice(s, func(i int, e string) {
		mu.Lock()
		defer mu.Unlock()
		res[i] = e
	})
	assert.SliceEqual(t, res, s)
}

func TestMap(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	mu := new(sync.Mutex)
	res := make(map[string]int)
	Map(m, func(k string, v int) {
		mu.Lock()
		defer mu.Unlock()
		res[k] = v
	})
	assert.MapEqual(t, res, m)
}
