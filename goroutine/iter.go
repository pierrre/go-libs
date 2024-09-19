package goroutine

import (
	"context"
	"iter"

	"github.com/pierrre/go-libs/panichandle"
)

// Iter runs a function on an iterator with concurrent workers.
func Iter[In, Out any](ctx context.Context, in iter.Seq[In], workers int, f func(context.Context, In) Out) iter.Seq[Out] {
	return func(yield func(Out) bool) {
		ctx, cancel := context.WithCancel(ctx) //nolint:govet // Shadows ctx.
		defer cancel()
		inCh := make(chan In)
		go func() {
			defer close(inCh)
			for inV := range in {
				inCh <- inV
				select {
				case <-ctx.Done():
					return
				default:
				}
			}
		}()
		stoppedOutIter := make(chan struct{})
		outCh := make(chan Out)
		wg := waitGroupPool.Get()
		wg.Add(workers)
		for range workers {
			go func() {
				defer wg.Done()
				for inV := range inCh {
					func() {
						defer panichandle.Recover(ctx)
						outV := f(ctx, inV)
						select {
						case outCh <- outV:
						case <-stoppedOutIter:
						}
					}()
				}
			}()
		}
		go func() {
			wg.Wait()
			waitGroupPool.Put(wg)
			close(outCh)
		}()
		defer cancel()
		defer close(stoppedOutIter)
		for v := range outCh {
			if !yield(v) {
				return
			}
		}
	}
}

// ValErr is a value with an error.
type ValErr[T any] struct {
	Val T
	Err error
}

// WithError transforms a function that returns an error into a function that returns a [ValErr].
func WithError[In, Out any](f func(context.Context, In) (Out, error)) func(context.Context, In) ValErr[Out] {
	return func(ctx context.Context, v In) ValErr[Out] {
		out, err := f(ctx, v)
		return ValErr[Out]{
			Val: out,
			Err: err,
		}
	}
}
