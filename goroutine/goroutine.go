// Package goroutine helps to manage goroutines safely.
//
// It recovers panic with [panichandle].
package goroutine

import (
	"context"
	"sync"

	"github.com/pierrre/go-libs/panichandle"
)

// Start executes a function in a new goroutine.
func Start(ctx context.Context, f func(ctx context.Context)) {
	go func() {
		defer panichandle.Recover(ctx)
		f(ctx)
	}()
}

// Wait executes a function in a new goroutine.
// It returns a function that blocks until the goroutine is terminated.
// The caller must call this function.
func Wait(ctx context.Context, f func(ctx context.Context)) (wait func()) {
	wg := new(sync.WaitGroup)
	WaitGroup(ctx, wg, f)
	return wg.Wait
}

// CancelWait executes a function in a new goroutine.
// It returns a function that cancels the context and blocks until the goroutine is terminated.
// The caller must call this function.
func CancelWait(ctx context.Context, f func(ctx context.Context)) (cancelWait func()) {
	ctx, cancel := context.WithCancel(ctx)
	wg := new(sync.WaitGroup)
	WaitGroup(ctx, wg, f)
	return func() {
		cancel()
		wg.Wait()
	}
}

// WaitGroup executes a function in a new goroutine with a [sync.WaitGroup].
// It calls [sync.WaitGroup.Add] before starting it, and [sync.WaitGroup.Done] when the goroutine is terminated.
func WaitGroup(ctx context.Context, wg *sync.WaitGroup, f func(ctx context.Context)) {
	wg.Add(1)
	go func() {
		defer panichandle.Recover(ctx)
		defer wg.Done()
		f(ctx)
	}()
}

// N executes a function with multiple goroutines.
// It blocks until all goroutines are terminated.
func N(ctx context.Context, n int, f func(ctx context.Context)) {
	wg := new(sync.WaitGroup)
	for range n {
		WaitGroup(ctx, wg, f)
	}
	wg.Wait()
}
