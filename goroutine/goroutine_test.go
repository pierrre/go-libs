package goroutine

import (
	"context"
	"fmt"
	"runtime"
	"slices"
	"sync/atomic"
	"testing"
	"testing/synctest"

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

func ExampleStartN() {
	ctx := context.Background()
	wait := StartN(ctx, 3, func(ctx context.Context, i int) {
		fmt.Println(i)
	})
	wait.Wait()
	// Unordered output:
	// 0
	// 1
	// 2
}

func ExampleStartNWithCancel() {
	ctx := context.Background()
	cancelWait := StartNWithCancel(ctx, 3, func(ctx context.Context, i int) {
		fmt.Println(i)
		<-ctx.Done()
	})
	cancelWait.Wait()
	// Unordered output:
	// 0
	// 1
	// 2
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
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		var called int64
		wait := Start(ctx, func(ctx context.Context) {
			atomic.AddInt64(&called, 1)
		})
		wait.Wait()
		assert.Equal(t, called, 1)
	})
}

func TestStartAllocs(t *testing.T) {
	ctx := t.Context()
	assertauto.AllocsPerRun(t, 100, func() {
		wait := Start(ctx, func(ctx context.Context) {})
		wait.Wait()
	})
}

func TestStartPanic(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		var called atomic.Int64
		wait := Start(ctx, func(ctx context.Context) {
			called.Add(1)
			panic("panic")
		})
		assert.Panics(t, func() {
			wait.Wait()
		})
	})
}

func TestStartGoexit(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		normalReturn := false
		done := make(chan struct{})
		go func() {
			defer close(done)
			wait := Start(ctx, func(ctx context.Context) {
				runtime.Goexit()
			})
			assert.NotPanics(t, func() {
				wait.Wait()
			})
			normalReturn = true
		}()
		<-done
		assert.False(t, normalReturn)
	})
}

func TestStartGoexitAndPanic(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		normalReturn := false
		done := make(chan struct{})
		go func() {
			defer close(done)
			wait := Start(ctx, func(ctx context.Context) {
				defer panic("panic")
				runtime.Goexit()
			})
			assert.Panics(t, func() {
				wait.Wait()
			})
			normalReturn = true
		}()
		<-done
		assert.False(t, normalReturn)
	})
}

func TestStartNoTerminationPropagation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		ctx = WithTerminationPropagation(ctx, false)
		wait := Start(ctx, func(ctx context.Context) {
			runtime.Goexit()
		})
		wait.Wait()
	})
}

func BenchmarkStart(b *testing.B) {
	ctx := b.Context()
	for b.Loop() {
		wait := Start(ctx, func(ctx context.Context) {})
		wait.Wait()
	}
}

func TestStartWithCancel(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		var called int64
		cancelWait := StartWithCancel(ctx, func(ctx context.Context) {
			atomic.AddInt64(&called, 1)
			<-ctx.Done()
		})
		cancelWait.Wait()
		assert.Equal(t, called, 1)
	})
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

func TestStartNWithCancel(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		called := make([]int64, 10)
		cancelWait := StartNWithCancel(ctx, 10, func(ctx context.Context, i int) {
			atomic.AddInt64(&called[i], 1)
			<-ctx.Done()
		})
		cancelWait.Wait()
		expected := slices.Repeat([]int64{1}, 10)
		assert.SliceEqual(t, called, expected)
	})
}

func TestRunN(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		called := make([]int64, 10)
		RunN(ctx, 10, func(ctx context.Context, i int) {
			atomic.AddInt64(&called[i], 1)
		})
		expected := slices.Repeat([]int64{1}, 10)
		assert.SliceEqual(t, called, expected)
	})
}

func TestRunNAllocs(t *testing.T) {
	ctx := t.Context()
	assertauto.AllocsPerRun(t, 100, func() {
		RunN(ctx, 10, func(ctx context.Context, _ int) {})
	})
}

func TestRunNContextCancel(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		ctx, cancel := context.WithCancel(ctx)
		var count atomic.Int64
		RunN(ctx, 10, func(ctx context.Context, _ int) {
			if count.Add(1) == 5 {
				cancel()
			}
			<-ctx.Done()
		})
	})
}

func TestRunNZero(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		var called int64
		RunN(ctx, 0, func(ctx context.Context, _ int) {
			atomic.AddInt64(&called, 1)
		})
		assert.Equal(t, called, 0)
	})
}

func TestRunNNegativePanic(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		assert.Panics(t, func() {
			RunN(ctx, -10, func(ctx context.Context, _ int) {})
		})
	})
}

func TestRunNPanic(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		var counter atomic.Int64
		assert.Panics(t, func() {
			RunN(ctx, 10, func(ctx context.Context, _ int) {
				id := counter.Add(1)
				if id == 1 {
					panic("panic")
				}
				<-ctx.Done()
			})
		})
	})
}

func TestRunNPanicAll(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		assert.Panics(t, func() {
			RunN(ctx, 10, func(ctx context.Context, _ int) {
				panic("panic")
			})
		})
	})
}

func TestRunNGoexit(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		normalReturn := false
		done := make(chan struct{})
		go func() {
			defer close(done)
			assert.NotPanics(t, func() {
				RunN(ctx, 10, func(ctx context.Context, _ int) {
					runtime.Goexit()
				})
			})
			normalReturn = true
		}()
		<-done
		assert.False(t, normalReturn)
	})
}

func TestRunNGoexitAndPanic(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		normalReturn := false
		done := make(chan struct{})
		go func() {
			defer close(done)
			assert.Panics(t, func() {
				RunN(ctx, 10, func(ctx context.Context, _ int) {
					defer panic("panic")
					runtime.Goexit()
				})
			})
			normalReturn = true
		}()
		<-done
		assert.False(t, normalReturn)
	})
}

func TestRunNNoTerminationPropagation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		ctx = WithTerminationPropagation(ctx, false)
		RunN(ctx, 10, func(ctx context.Context, _ int) {
			runtime.Goexit()
		})
	})
}

func BenchmarkRunN(b *testing.B) {
	ctx := b.Context()
	for b.Loop() {
		RunN(ctx, 10, func(ctx context.Context, _ int) {})
	}
}

func TestInitTerminationPropagationEnabledDefault(t *testing.T) {
	enabled := initTerminationPropagationEnabled()
	assert.True(t, enabled)
}

func TestInitTerminationPropagationEnabledEnvTrue(t *testing.T) {
	t.Setenv(terminationPropagationEnabledEnv, "true")
	enabled := initTerminationPropagationEnabled()
	assert.True(t, enabled)
}

func TestInitTerminationPropagationEnabledEnvFalse(t *testing.T) {
	t.Setenv(terminationPropagationEnabledEnv, "false")
	enabled := initTerminationPropagationEnabled()
	assert.False(t, enabled)
}

func TestInitTerminationPropagationEnabledEnvPanic(t *testing.T) {
	t.Setenv(terminationPropagationEnabledEnv, "invalid")
	assert.Panics(t, func() {
		initTerminationPropagationEnabled()
	})
}
