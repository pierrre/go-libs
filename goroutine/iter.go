package goroutine

import (
	"context"
	"iter"

	"github.com/pierrre/go-libs/panichandle"
)

// Iter runs a function for each values of an input iterator with concurrent workers.
// It returns an iterator with the output values.
//
// If the context is canceled, it stops reading values from the input iterator.
// If the caller stops iterating the output iterator, the context is canceled.
func Iter[In, Out any](ctx context.Context, in iter.Seq[In], workers int, f func(context.Context, In) Out) iter.Seq[Out] {
	return func(yield func(Out) bool) {
		iterSeq(ctx, in, workers, f, yield)
	}
}

func iterSeq[In, Out any](ctx context.Context, in iter.Seq[In], workers int, f func(context.Context, In) Out, yield func(Out) bool) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	workers = max(workers, 1)
	inCh := make(chan In)
	outCh := make(chan Out)
	go iterProducer(ctx, in, inCh)
	iterWorkers(ctx, workers, f, inCh, outCh)
	iterConsumer(cancel, outCh, yield)
}

func iterProducer[In any](ctx context.Context, in iter.Seq[In], inCh chan<- In) {
	defer close(inCh)
	for inV := range in {
		inCh <- inV
		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func iterWorkers[In, Out any](ctx context.Context, workers int, f func(context.Context, In) Out, inCh <-chan In, outCh chan<- Out) {
	wg := waitGroupPool.Get()
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			iterWorker(ctx, f, inCh, outCh)
		}()
	}
	go func() {
		wg.Wait()
		waitGroupPool.Put(wg)
		close(outCh)
	}()
}

func iterWorker[In, Out any](ctx context.Context, f func(context.Context, In) Out, inCh <-chan In, outCh chan<- Out) {
	for inV := range inCh {
		func() {
			defer panichandle.Recover(ctx)
			outV := f(ctx, inV)
			outCh <- outV
		}()
	}
}

func iterConsumer[Out any](cancel context.CancelFunc, outCh <-chan Out, yield func(Out) bool) {
	stopped := false
	for outV := range outCh {
		if stopped {
			continue
		}
		if !yield(outV) {
			cancel()
			stopped = true
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
	return func(ctx context.Context, inV In) ValErr[Out] {
		outV, err := f(ctx, inV)
		return ValErr[Out]{
			Val: outV,
			Err: err,
		}
	}
}
