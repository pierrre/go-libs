package goroutine

import (
	"context"
	"iter"
	"sync"

	"github.com/pierrre/go-libs/panichandle"
)

// Iter runs a function on an iterator with concurrent workers.
func Iter[In, Out any](ctx context.Context, in iter.Seq[In], workers int, f func(context.Context, In) Out) iter.Seq[Out] {
	return func(yield func(Out) bool) {
		iterSeq(ctx, in, workers, f, yield)
	}
}

func iterSeq[In, Out any](ctx context.Context, in iter.Seq[In], workers int, f func(context.Context, In) Out, yield func(Out) bool) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	inCh := make(chan In)
	outCh := make(chan Out)
	workersWg := new(sync.WaitGroup)
	defer workersWg.Wait()
	consumerStopped := make(chan struct{})
	go iterProducer(ctx, in, inCh)
	iterWorkers(ctx, workersWg, workers, f, inCh, outCh, consumerStopped)
	iterConsumer(cancel, consumerStopped, outCh, yield)
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

func iterWorkers[In, Out any](ctx context.Context, wg *sync.WaitGroup, workers int, f func(context.Context, In) Out, inCh <-chan In, outCh chan<- Out, stoppedOutIter <-chan struct{}) {
	wg.Add(workers)
	for range workers {
		go iterWorker(ctx, wg, f, inCh, outCh, stoppedOutIter)
	}
	go func() {
		wg.Wait()
		close(outCh)
	}()
}

func iterWorker[In, Out any](ctx context.Context, wg *sync.WaitGroup, f func(context.Context, In) Out, inCh <-chan In, outCh chan<- Out, consumerStopped <-chan struct{}) {
	defer wg.Done()
	for inV := range inCh {
		func() {
			defer panichandle.Recover(ctx)
			outV := f(ctx, inV)
			select {
			case outCh <- outV:
			case <-consumerStopped:
			}
		}()
	}
}

func iterConsumer[Out any](cancel context.CancelFunc, consumerStopped chan<- struct{}, outCh <-chan Out, yield func(Out) bool) {
	defer func() {
		cancel()
		close(consumerStopped)
	}()
	for outV := range outCh {
		if !yield(outV) {
			return
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
