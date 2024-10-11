package goroutine

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/pierrre/assert"
	"github.com/pierrre/assert/assertauto"
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
	assertauto.AllocsPerRun(t, 100, func() {
		Start(ctx, f)
		<-done
	})
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
	assertauto.AllocsPerRun(t, 100, func() {
		wait := Wait(ctx, func(ctx context.Context) {})
		wait()
	})
}

func BenchmarkWait(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()
	for range b.N {
		wait := Wait(ctx, func(ctx context.Context) {})
		wait()
	}
}

func TestCancelWait(t *testing.T) {
	ctx := context.Background()
	var called int64
	cancelWait := CancelWait(ctx, func(ctx context.Context) {
		atomic.AddInt64(&called, 1)
		<-ctx.Done()
	})
	cancelWait()
	assert.Equal(t, called, 1)
}

func TestCancelWaitAllocs(t *testing.T) {
	ctx := context.Background()
	assertauto.AllocsPerRun(t, 100, func() {
		cancelWait := CancelWait(ctx, func(ctx context.Context) {
			<-ctx.Done()
		})
		cancelWait()
	})
}

func BenchmarkCancelWait(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()
	for range b.N {
		cancelWait := CancelWait(ctx, func(ctx context.Context) {
			<-ctx.Done()
		})
		cancelWait()
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
	assertauto.AllocsPerRun(t, 100, func() {
		WaitGroup(ctx, wg, func(ctx context.Context) {})
		wg.Wait()
	})
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
	assertauto.AllocsPerRun(t, 100, func() {
		N(ctx, count, func(ctx context.Context) {})
	})
}

func BenchmarkN(b *testing.B) {
	ctx := context.Background()
	count := 10
	b.ResetTimer()
	for range b.N {
		N(ctx, count, func(ctx context.Context) {})
	}
}
