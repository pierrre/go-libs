// Package goroutine helps to manage goroutines safely.
//
// It recovers panic with [panichandle].
package goroutine

import (
	"context"
	"sync"

	"github.com/pierrre/go-libs/panichandle"
)

// Go executes a function in a new goroutine.
func Go(ctx context.Context, f func(ctx context.Context)) {
	go func() {
		defer panichandle.Recover(ctx)
		f(ctx)
	}()
}

// GoWait executes a function in a new goroutine.
//
// It returns a function that blocks until the goroutine is terminated.
// The caller must call this function.
func GoWait(ctx context.Context, f func(ctx context.Context)) (wait func()) {
	ch := make(chan struct{})
	Go(ctx, func(ctx context.Context) {
		defer close(ch)
		f(ctx)
	})
	return func() {
		<-ch
	}
}

// WaitGroup executes a function in a new goroutine with a [sync.WaitGroup].
// It calls [sync.WaitGroup.Add] before starting it, and [sync.WaitGroup.Done] when the goroutine is terminated.
func WaitGroup(ctx context.Context, wg *sync.WaitGroup, f func(ctx context.Context)) {
	wg.Add(1)
	Go(ctx, func(ctx context.Context) {
		defer wg.Done()
		f(ctx)
	})
}

// N executes a function with multiple goroutines.
// It blocks until all goroutines are terminated.
func N(ctx context.Context, n int, f func(ctx context.Context, i int)) {
	wg := new(sync.WaitGroup)
	for i := 0; i < n; i++ {
		i := i
		WaitGroup(ctx, wg, func(ctx context.Context) {
			f(ctx, i)
		})
	}
	wg.Wait()
}

// Slice executes a function with a different goroutine for each element of the slice.
// It blocks until all goroutines are terminated.
func Slice[S ~[]E, E any](ctx context.Context, s S, f func(ctx context.Context, i int, e E)) {
	wg := new(sync.WaitGroup)
	for i, e := range s {
		i, e := i, e
		WaitGroup(ctx, wg, func(ctx context.Context) {
			f(ctx, i, e)
		})
	}
	wg.Wait()
}

// Map executes a function with a different goroutine for each element of the map.
// It blocks until all goroutines are terminated.
func Map[M ~map[K]V, K comparable, V any](ctx context.Context, m M, f func(ctx context.Context, k K, v V)) {
	wg := new(sync.WaitGroup)
	for k, v := range m {
		k, v := k, v
		WaitGroup(ctx, wg, func(ctx context.Context) {
			f(ctx, k, v)
		})
	}
	wg.Wait()
}
