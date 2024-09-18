// Package goroutine helps to manage goroutines safely.
//
// It recovers panic with [panichandle].
package goroutine

import (
	"context"
	"sync"

	"github.com/pierrre/go-libs/panichandle"
	"github.com/pierrre/go-libs/syncutil"
)

// Start executes a function in a new goroutine.
func Start(ctx context.Context, f func(ctx context.Context)) {
	go func() {
		defer panichandle.Recover(ctx)
		f(ctx)
	}()
}

// Wait executes a function in a new goroutine and allows to wait until it terminates.
// It returns a function that blocks until the goroutine is terminated.
// The caller must call this function.
func Wait(ctx context.Context, f func(ctx context.Context)) (wait func()) {
	wg := new(sync.WaitGroup)
	WaitGroup(ctx, wg, f)
	return wg.Wait
}

// WaitGroup executes a function in a new goroutine with a [sync.WaitGroup].
// It calls [sync.WaitGroup.Add] before starting it, and [sync.WaitGroup.Done] when the goroutine is terminated.
func WaitGroup(ctx context.Context, wg *sync.WaitGroup, f func(ctx context.Context)) {
	wg.Add(1)
	Start(ctx, func(ctx context.Context) {
		defer wg.Done()
		f(ctx)
	})
}

// N executes a function with multiple goroutines.
// It blocks until all goroutines are terminated.
func N(ctx context.Context, n int, f func(ctx context.Context)) {
	wg := waitGroupPool.Get()
	for range n {
		WaitGroup(ctx, wg, f)
	}
	wg.Wait()
	waitGroupPool.Put(wg)
}

var waitGroupPool = syncutil.Pool[*sync.WaitGroup]{
	New: func() *sync.WaitGroup {
		return new(sync.WaitGroup)
	},
}
