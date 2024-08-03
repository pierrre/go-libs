package goroutine

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/pierrre/assert"
)

func TestStart(t *testing.T) {
	ctx := context.Background()
	var called int64
	done := make(chan struct{})
	Start(ctx, func(ctx context.Context) {
		atomic.AddInt64(&called, 1)
		close(done)
	})
	<-done
	assert.Equal(t, called, 1)
}

func TestStartAllocs(t *testing.T) {
	ctx := context.Background()
	done := make(chan struct{})
	assert.AllocsPerRun(t, 100, func() {
		Start(ctx, func(ctx context.Context) {
			done <- struct{}{}
		})
		<-done
	}, 2)
}

func TestWait(t *testing.T) {
	ctx := context.Background()
	var called int64
	wait := Wait(ctx, func(ctx context.Context) {
		atomic.AddInt64(&called, 1)
	})
	wait()
	assert.Equal(t, called, 1)
}

func TestWaitAllocs(t *testing.T) {
	ctx := context.Background()
	assert.AllocsPerRun(t, 100, func() {
		wait := Wait(ctx, func(ctx context.Context) {})
		wait()
	}, 4)
}

func TestWaitGroup(t *testing.T) {
	ctx := context.Background()
	wg := new(sync.WaitGroup)
	var called int64
	WaitGroup(ctx, wg, func(ctx context.Context) {
		atomic.AddInt64(&called, 1)
	})
	wg.Wait()
	assert.Equal(t, called, 1)
}

func TestWaitGroupAllocs(t *testing.T) {
	ctx := context.Background()
	wg := new(sync.WaitGroup)
	assert.AllocsPerRun(t, 100, func() {
		WaitGroup(ctx, wg, func(ctx context.Context) {})
		wg.Wait()
	}, 2)
}

func TestN(t *testing.T) {
	ctx := context.Background()
	count := 10
	mu := new(sync.Mutex)
	is := make(map[int]struct{})
	N(ctx, count, func(ctx context.Context, i int) {
		mu.Lock()
		defer mu.Unlock()
		is[i] = struct{}{}
	})
	isExpected := make(map[int]struct{})
	for i := range count {
		isExpected[i] = struct{}{}
	}
	assert.MapEqual(t, is, isExpected)
}

func TestSlice(t *testing.T) {
	ctx := context.Background()
	s := []string{"a", "b", "c"}
	mu := new(sync.Mutex)
	res := make([]string, len(s))
	Slice(ctx, s, func(ctx context.Context, i int, e string) {
		mu.Lock()
		defer mu.Unlock()
		res[i] = e
	})
	assert.SliceEqual(t, res, s)
}

func TestMap(t *testing.T) {
	ctx := context.Background()
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	mu := new(sync.Mutex)
	res := make(map[string]int)
	Map(ctx, m, func(ctx context.Context, k string, v int) {
		mu.Lock()
		defer mu.Unlock()
		res[k] = v
	})
	assert.MapEqual(t, res, m)
}
