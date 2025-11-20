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

	"github.com/pierrre/go-libs/funcutil"
)

// Waiter is an interface for waiting for goroutines to finish.
type Waiter interface {
	// Wait blocks until all goroutines are finished.
	//
	// It propagates panics or calls to [runtime.Goexit] from the goroutines to the caller.
	// The propagation behavior can be configured with [WithTerminationPropagation] (enabled by default).
	Wait()
}

type waiterFunc func()

func (f waiterFunc) Wait() {
	f()
}

// Start executes a function in a new goroutine.
// The caller must call the returned [Waiter].
func Start(ctx context.Context, f func(ctx context.Context)) Waiter {
	propagateTermination := isTerminationPropagationEnabled(ctx)
	res := new(startResult)
	res.wg.Add(1)
	go func() {
		if propagateTermination {
			funcutil.Call(
				func() {
					f(ctx)
				},
				func(goexit bool, panicErr error) {
					res.goexit = goexit
					res.panicErr = panicErr
					res.wg.Done()
				},
			)
		} else {
			defer res.wg.Done()
			f(ctx)
		}
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
	propagateTermination := isTerminationPropagationEnabled(ctx)
	ctx, cancel := context.WithCancel(ctx)
	res := &startNResult{
		cancel: cancel,
	}
	res.wg.Add(n)
	for i := range n {
		go func() {
			if propagateTermination {
				funcutil.Call(
					func() {
						f(ctx, i)
					},
					func(goexit bool, panicErr error) {
						res.mu.Lock()
						if goexit {
							cancel()
							res.goexit = true
						}
						if panicErr != nil {
							cancel()
							res.panicErrs = append(res.panicErrs, panicErr)
						}
						res.mu.Unlock()
						res.wg.Done()
					},
				)
			} else {
				defer res.wg.Done()
				f(ctx, i)
			}
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

type terminationPropagationContextKey struct{}

// WithTerminationPropagation configures whether termination propagation is enabled in the [context.Context].
// It determines if abnormal termination ([panic] or [runtime.Goexit]) in goroutines should be propagated to the caller.
func WithTerminationPropagation(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, terminationPropagationContextKey{}, enabled)
}

func isTerminationPropagationEnabled(ctx context.Context) bool {
	v, ok := ctx.Value(terminationPropagationContextKey{}).(bool)
	return !ok || v
}
