package goroutine

import (
	"context"
	"fmt"
	"runtime"
	"slices"
	"sync/atomic"
	"testing"

	"github.com/pierrre/assert"
	"github.com/pierrre/assert/assertauto"
)

func ExampleStart() {
	ctx := context.Background()
	wait := Start(ctx, func(ctx context.Context) {
		fmt.Println("a")
	})
	wait.Wait()
	// Output:
	// a
}

func ExampleStartWithCancel() {
	ctx := context.Background()
	cancelWait := StartWithCancel(ctx, func(ctx context.Context) {
		fmt.Println("a")
		<-ctx.Done()
	})
	cancelWait.Wait()
	// Output:
	// a
}

func ExampleRunN() {
	ctx := context.Background()
	RunN(ctx, 3, func(ctx context.Context, i int) {
		fmt.Println(i)
	})
	// Unordered output:
	// 0
	// 1
	// 2
}

func TestStart(t *testing.T) {
	ctx := t.Context()
	var called int64
	wait := Start(ctx, func(ctx context.Context) {
		atomic.AddInt64(&called, 1)
	})
	wait.Wait()
	assert.Equal(t, called, 1)
}

func TestStartAllocs(t *testing.T) {
	ctx := t.Context()
	assertauto.AllocsPerRun(t, 100, func() {
		wait := Start(ctx, func(ctx context.Context) {})
		wait.Wait()
	})
}

func TestStartPanic(t *testing.T) {
	ctx := t.Context()
	var called int64
	wait := Start(ctx, func(ctx context.Context) {
		atomic.AddInt64(&called, 1)
		panic("panic")
	})
	assert.Panics(t, func() {
		wait.Wait()
	})
}

func TestStartGoexit(t *testing.T) {
	ctx := t.Context()
	normalReturn := false
	recovered := false
	done := make(chan struct{})
	go func() {
		defer func() {
			r := recover()
			if r != nil {
				recovered = true
			}
			close(done)
		}()
		wait := Start(ctx, func(ctx context.Context) {
			runtime.Goexit()
		})
		wait.Wait()
		normalReturn = true
	}()
	<-done
	assert.False(t, normalReturn)
	assert.False(t, recovered)
}

func BenchmarkStart(b *testing.B) {
	ctx := b.Context()
	for b.Loop() {
		wait := Start(ctx, func(ctx context.Context) {})
		wait.Wait()
	}
}

func TestStartWithCancel(t *testing.T) {
	ctx := t.Context()
	var called int64
	cancelWait := StartWithCancel(ctx, func(ctx context.Context) {
		atomic.AddInt64(&called, 1)
		<-ctx.Done()
	})
	cancelWait.Wait()
	assert.Equal(t, called, 1)
}

func TestStartWithCancelAllocs(t *testing.T) {
	ctx := t.Context()
	assertauto.AllocsPerRun(t, 100, func() {
		cancelWait := StartWithCancel(ctx, func(ctx context.Context) {
			<-ctx.Done()
		})
		cancelWait.Wait()
	})
}

func BenchmarkStartWithCancel(b *testing.B) {
	ctx := b.Context()
	for b.Loop() {
		cancelWait := StartWithCancel(ctx, func(ctx context.Context) {
			<-ctx.Done()
		})
		cancelWait.Wait()
	}
}

func TestRunN(t *testing.T) {
	ctx := t.Context()
	called := make([]int64, 10)
	RunN(ctx, 10, func(ctx context.Context, i int) {
		atomic.AddInt64(&called[i], 1)
	})
	expected := slices.Repeat([]int64{1}, 10)
	assert.SliceEqual(t, called, expected)
}

func TestRunNAllocs(t *testing.T) {
	ctx := t.Context()
	assertauto.AllocsPerRun(t, 100, func() {
		RunN(ctx, 10, func(ctx context.Context, _ int) {})
	})
}

func TestRunNContextCancel(t *testing.T) {
	ctx := t.Context()
	ctx, cancel := context.WithCancel(ctx)
	var count int64
	RunN(ctx, 10, func(ctx context.Context, _ int) {
		if atomic.AddInt64(&count, 1) == 5 {
			cancel()
		}
		<-ctx.Done()
	})
}

func TestRunNZero(t *testing.T) {
	ctx := t.Context()
	var called int64
	RunN(ctx, 0, func(ctx context.Context, _ int) {
		atomic.AddInt64(&called, 1)
	})
	assert.Equal(t, called, 0)
}

func TestRunNNegativePanic(t *testing.T) {
	ctx := t.Context()
	assert.Panics(t, func() {
		RunN(ctx, -10, func(ctx context.Context, _ int) {})
	})
}

func TestRunNPanic(t *testing.T) {
	ctx := t.Context()
	var counter int64
	assert.Panics(t, func() {
		RunN(ctx, 10, func(ctx context.Context, _ int) {
			id := atomic.AddInt64(&counter, 1)
			if id == 1 {
				panic("panic")
			}
			<-ctx.Done()
		})
	})
}

func TestRunNPanicAll(t *testing.T) {
	ctx := t.Context()
	assert.Panics(t, func() {
		RunN(ctx, 10, func(ctx context.Context, _ int) {
			panic("panic")
		})
	})
}

func TestRunNGoexit(t *testing.T) {
	ctx := t.Context()
	normalReturn := false
	recovered := false
	done := make(chan struct{})
	go func() {
		defer func() {
			r := recover()
			if r != nil {
				recovered = true
			}
			close(done)
		}()
		RunN(ctx, 10, func(ctx context.Context, _ int) {
			runtime.Goexit()
		})
		normalReturn = true
	}()
	<-done
	assert.False(t, normalReturn)
	assert.False(t, recovered)
}

func BenchmarkRunN(b *testing.B) {
	ctx := b.Context()
	for b.Loop() {
		RunN(ctx, 10, func(ctx context.Context, _ int) {})
	}
}
