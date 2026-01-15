// Package goroutine helps to manage goroutines safely.
//
//   - Start goroutines: [Start], [StartWithCancel], [RunN].
//   - Process iterators: [Iter], [IterOrdered], [WithError].
//   - Process slices: [Slice], [SliceError]
//   - Process maps: [Map], [MapError].
//   - Run services: [Services].
package goroutine

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/pierrre/go-libs/funcutil"
)

// Waiter is an interface for waiting for goroutines to finish.
type Waiter interface {
	// Wait blocks until all goroutines are finished.
	//
	// It propagates panics or calls to [runtime.Goexit] from the goroutines to the caller.
	// The propagation behavior can be configured with [TerminationPropagationEnabled] or [WithTerminationPropagation].
	Wait()
}

type waiterFunc func()

func (f waiterFunc) Wait() { f() }

// Start executes a function in a new goroutine.
// The caller must call the returned [Waiter].
func Start(ctx context.Context, f func(ctx context.Context)) Waiter {
	if !isTerminationPropagationEnabled(ctx) {
		wg := new(sync.WaitGroup)
		wg.Go(func() { f(ctx) })
		return wg
	}
	res := new(startResult)
	res.wg.Add(1)
	go func() {
		funcutil.Call(
			func() { f(ctx) },
			func(goexit bool, panicErr error) {
				res.goexit = goexit
				res.panicErr = panicErr
				res.wg.Done()
			},
		)
	}()
	return res
}

type startResult struct {
	wg       sync.WaitGroup
	goexit   bool
	panicErr error
}

func (res *startResult) Wait() {
	res.wg.Wait()
	if res.goexit {
		runtime.Goexit()
	}
	if res.panicErr != nil {
		panic(res.panicErr)
	}
}

// StartWithCancel executes a function in a new goroutine and allows the caller to cancel it with the [Waiter].
// The caller must call the returned [Waiter].
func StartWithCancel(ctx context.Context, f func(ctx context.Context)) Waiter {
	ctx, cancel := context.WithCancel(ctx)
	wait := Start(ctx, f)
	return waiterFunc(func() {
		cancel()
		wait.Wait()
	})
}

func startN(ctx context.Context, n int, f func(ctx context.Context, i int)) Waiter {
	if !isTerminationPropagationEnabled(ctx) {
		wg := new(sync.WaitGroup)
		for i := range n {
			wg.Go(func() { f(ctx, i) })
		}
		return wg
	}
	res := new(startNResult)
	ctx, res.cancel = context.WithCancel(ctx)
	res.wg.Add(n)
	for i := range n {
		go func() {
			defer res.wg.Done()
			funcutil.Call(
				func() { f(ctx, i) },
				func(goexit bool, panicErr error) {
					res.mu.Lock()
					if goexit {
						res.cancel()
						res.goexit = true
					}
					if panicErr != nil {
						res.cancel()
						res.panicErrs = append(res.panicErrs, panicErr)
					}
					res.mu.Unlock()
				},
			)
		}()
	}
	return res
}

type startNResult struct {
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	mu        sync.Mutex
	goexit    bool
	panicErrs []error
}

func (res *startNResult) Wait() {
	res.wg.Wait()
	res.cancel()
	if res.goexit {
		runtime.Goexit()
	}
	if len(res.panicErrs) > 0 {
		err := res.panicErrs[0]
		if len(res.panicErrs) > 1 {
			err = errors.Join(res.panicErrs...)
		}
		panic(err)
	}
}

// RunN executes a function with multiple goroutines.
// The i parameter is the index of the goroutine (from 0 to n-1).
// It blocks until all goroutines are terminated (see [Waiter.Wait]).
func RunN(ctx context.Context, n int, f func(ctx context.Context, i int)) {
	res := startN(ctx, n, f)
	res.Wait()
}

// TerminationPropagationEnabled indicates whether termination propagation is enabled (true by default).
// It determines if abnormal termination ([panic] or [runtime.Goexit]) in goroutines should be propagated to the caller.
// It can be overridden per goroutine with [WithTerminationPropagation].
var TerminationPropagationEnabled atomic.Bool

func init() {
	TerminationPropagationEnabled.Store(true)
}

type terminationPropagationContextKey struct{}

// WithTerminationPropagation configures termination propagation for goroutines started with the given [context.Context].
func WithTerminationPropagation(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, terminationPropagationContextKey{}, enabled)
}

func isTerminationPropagationEnabled(ctx context.Context) bool {
	v, ok := ctx.Value(terminationPropagationContextKey{}).(bool)
	if ok {
		return v
	}
	return TerminationPropagationEnabled.Load()
}
