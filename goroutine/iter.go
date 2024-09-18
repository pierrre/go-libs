package goroutine

import (
	"context"
	"iter"
)

// Func is a function that receives an In value and returns an Out value.
type Func[In, Out any] func(ctx context.Context, v In) Out

// Iter runs a function on an iterator with n concurrent workers.
func Iter[In, Out any](ctx context.Context, in iter.Seq[In], n int, f Func[In, Out]) iter.Seq[Out] {
	return func(yield func(Out) bool) {
		ctx, cancel := context.WithCancel(ctx) //nolint:govet // Shadows ctx.
		defer cancel()
		inCh := make(chan In)
		defer Wait(ctx, func(ctx context.Context) {
			defer close(inCh)
			for inV := range in {
				inCh <- inV
				select {
				case <-ctx.Done():
					return
				default:
				}
			}
		})()
		stoppedOutIter := make(chan struct{})
		outCh := make(chan Out)
		defer Wait(ctx, func(ctx context.Context) {
			defer close(outCh)
			N(ctx, n, func(ctx context.Context) {
				for inV := range inCh {
					outV := f(ctx, inV)
					select {
					case outCh <- outV:
					case <-stoppedOutIter:
					}
				}
			})
		})()
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

// ErrorFunc is a [Func] that also returns an error.
type ErrorFunc[In, Out any] func(ctx context.Context, v In) (Out, error)

// WithError transforms a [ErrorFunc] into a [Func].
func WithError[In, Out any](f ErrorFunc[In, Out]) Func[In, ValErr[Out]] {
	return func(ctx context.Context, v In) ValErr[Out] {
		out, err := f(ctx, v)
		return ValErr[Out]{
			Val: out,
			Err: err,
		}
	}
}
