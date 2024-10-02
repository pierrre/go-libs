package goroutine

import (
	"context"
	"iter"
	"sync"
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

// iterSeq runs the iterator logic.
//
// It is made up 3 components: producer, workers, and consumer.
// The producer reads the input iterator and sends the input values to the workers.
// The workers run the function for each input values with a goroutines pool, and send the output values to the consumer.
// The consumer receives the output values from the workers, and sends them to the output iterator.
func iterSeq[In, Out any](ctx context.Context, in iter.Seq[In], workers int, f func(context.Context, In) Out, yield func(Out) bool) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	inCh := make(chan In)
	outCh := make(chan Out)
	workersWg := new(sync.WaitGroup)
	defer workersWg.Wait()                 // Wait for all workers to finish.
	consumerStopped := make(chan struct{}) // Used to notify that the consumer has stopped.
	go iterProducer(ctx, in, inCh)
	iterWorkers(ctx, workersWg, workers, f, inCh, outCh, consumerStopped)
	iterConsumer(cancel, consumerStopped, outCh, yield)
}

func iterProducer[In any](ctx context.Context, in iter.Seq[In], inCh chan<- In) {
	defer close(inCh) // Notify the workers that there are no more input values.
	for inV := range in {
		inCh <- inV
		select {
		case <-ctx.Done():
			// The context is canceled, stop the producer.
			return
		default:
		}
	}
}

func iterWorkers[In, Out any](ctx context.Context, wg *sync.WaitGroup, workers int, f func(context.Context, In) Out, inCh <-chan In, outCh chan<- Out, consumerStopped <-chan struct{}) {
	// Start all workers.
	workers = max(workers, 1) // Ensure there is at least 1 worker.
	wg.Add(workers)
	for range workers {
		go iterWorker(ctx, wg, f, inCh, outCh, consumerStopped)
	}
	go func() {
		wg.Wait()    // Wait for all workers to finish.
		close(outCh) // Notify the consumer that there are no more output values.
	}()
}

func iterWorker[In, Out any](ctx context.Context, wg *sync.WaitGroup, f func(context.Context, In) Out, inCh <-chan In, outCh chan<- Out, consumerStopped <-chan struct{}) {
	defer wg.Done()         // Notify that the worker has finished.
	for inV := range inCh { // Receive the input value from the producer.
		func() {
			// TODO: call a "panic handler" when it's available.
			outV := f(ctx, inV)
			select {
			case outCh <- outV: // Send the output value to the consumer.
			case <-consumerStopped: // The consumer has stopped, discard the output value.
			}
		}()
	}
}

func iterConsumer[Out any](cancel context.CancelFunc, consumerStopped chan<- struct{}, outCh <-chan Out, yield func(Out) bool) {
	defer func() {
		cancel()               // Notify the producer to stop.
		close(consumerStopped) // Notify the workers that the consumer has stopped.
	}()
	for outV := range outCh { // Receive the output value from the workers.
		if !yield(outV) { // Send the output value to the output iterator.
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
	return func(ctx context.Context, inV In) ValErr[Out] {
		outV, err := f(ctx, inV)
		return ValErr[Out]{
			Val: outV,
			Err: err,
		}
	}
}
