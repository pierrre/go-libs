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
	f := func(_ context.Context) {
		atomic.AddInt64(&called, 1)
		done <- struct{}{}
	}
	Start(ctx, f)
	<-done
	assert.Equal(t, called, 1)
}

func TestStartAllocs(t *testing.T) {
	ctx := context.Background()
	done := make(chan struct{})
	f := func(_ context.Context) {
		done <- struct{}{}
	}
	assert.AllocsPerRun(t, 100, func() {
		Start(ctx, f)
		<-done
	}, 1)
}

func BenchmarkStart(b *testing.B) {
	ctx := context.Background()
	done := make(chan struct{})
	f := func(ctx context.Context) {
		done <- struct{}{}
	}
	b.ResetTimer()
	for range b.N {
		Start(ctx, f)
		<-done
	}
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
	}, 3)
}

func BenchmarkWait(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()
	for range b.N {
		wait := Wait(ctx, func(ctx context.Context) {})
		wait()
	}
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

func BenchmarkWaitGroup(b *testing.B) {
	ctx := context.Background()
	wg := new(sync.WaitGroup)
	b.ResetTimer()
	for range b.N {
		WaitGroup(ctx, wg, func(ctx context.Context) {})
		wg.Wait()
	}
}

func TestN(t *testing.T) {
	ctx := context.Background()
	count := 10
	var called int64
	N(ctx, count, func(ctx context.Context) {
		atomic.AddInt64(&called, 1)
	})
	assert.Equal(t, called, int64(count))
}

func TestNAllocs(t *testing.T) {
	ctx := context.Background()
	count := 10
	assert.AllocsPerRun(t, 100, func() {
		N(ctx, count, func(ctx context.Context) {})
	}, 20)
}

func BenchmarkN(b *testing.B) {
	ctx := context.Background()
	count := 10
	b.ResetTimer()
	for range b.N {
		N(ctx, count, func(ctx context.Context) {})
	}
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

func BenchmarkSlice(b *testing.B) {
	ctx := context.Background()
	s := []string{"a", "b", "c"}
	b.ResetTimer()
	for range b.N {
		Slice(ctx, s, func(ctx context.Context, i int, e string) {})
	}
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

func BenchmarkMap(b *testing.B) {
	ctx := context.Background()
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	b.ResetTimer()
	for range b.N {
		Map(ctx, m, func(ctx context.Context, k string, v int) {})
	}
}
