package goroutine

import (
	"context"
	"errors"
	"iter"
	"maps"
	"reflect"
	"slices"
	"sync"

	"github.com/pierrre/go-libs/iterutil"
	"github.com/pierrre/go-libs/panichandle"
	"github.com/pierrre/go-libs/syncutil"
)

// Iter runs a function for each values of an input iterator with concurrent workers.
// It returns an iterator with the unordered output values.
// For an ordered version, see [IterOrdered].
//
// If the context is canceled, it stops reading values from the input iterator.
// If the caller stops iterating the output iterator, the context is canceled.
func Iter[In, Out any](ctx context.Context, in iter.Seq[In], workers int, f func(context.Context, In) Out) iter.Seq[Out] {
	return func(yield func(Out) bool) {
		iterSeq(ctx, in, workers, f, yield)
	}
}

// iterSeq implements the [Iter] logic.
func iterSeq[In, Out any](ctx context.Context, in iter.Seq[In], workers int, f func(context.Context, In) Out, yield func(Out) bool) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	workers = max(workers, 1)
	inCh := make(chan In)
	outCh := make(chan Out, workers) // The buffer helps to not block the workers if the code reading the output iterator is slow.
	go iterProducer(ctx, in, inCh)
	iterWorkers(ctx, workers, f, inCh, outCh)
	iterConsumer(cancel, outCh, yield)
}

// iterProducer reads values from the input iterator and sends them to the workers.
func iterProducer[In any](ctx context.Context, in iter.Seq[In], inCh chan<- In) {
	defer close(inCh) // Notify the workers that there are no more values to process.
	for inV := range in {
		inCh <- inV
		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

// iterWorkers starts the worker goroutines, and closes the output channel when all workers are done.
func iterWorkers[In, Out any](ctx context.Context, workers int, f func(context.Context, In) Out, inCh <-chan In, outCh chan<- Out) {
	wg := new(sync.WaitGroup)
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			iterWorker(ctx, f, inCh, outCh)
		}()
	}
	go func() {
		wg.Wait()
		close(outCh) // Notify the consumer that all workers are done.
	}()
}

// iterWorker reads the input value from the input channel, runs the function, and sends the output value to the output channel.
func iterWorker[In, Out any](ctx context.Context, f func(context.Context, In) Out, inCh <-chan In, outCh chan<- Out) {
	for inV := range inCh {
		func() {
			defer panichandle.Recover(ctx)
			outV := f(ctx, inV)
			outCh <- outV
		}()
	}
}

// iterConsumer reads the output values from the output channel and yields them to the output iterator.
func iterConsumer[Out any](cancel context.CancelFunc, outCh <-chan Out, yield func(Out) bool) {
	stopped := false
	for outV := range outCh {
		if stopped {
			continue // Drain the output channel if the output iterator was stopped.
		}
		if !yield(outV) {
			cancel() // Notify the producer that the output iterator was stopped.
			stopped = true
		}
	}
}

// IterOrdered is like [Iter] but it keeps the order of the output values.
func IterOrdered[In, Out any](ctx context.Context, in iter.Seq[In], workers int, f func(context.Context, In) Out) iter.Seq[Out] {
	return func(yield func(Out) bool) {
		iterOrdered(ctx, in, workers, f, yield)
	}
}

// iterOrdered implements the [IterOrdered] logic.
func iterOrdered[In, Out any](ctx context.Context, in iter.Seq[In], workers int, f func(context.Context, In) Out, yield func(Out) bool) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	workers = max(workers, 1)
	inCh := make(chan iterOrderedValue[In, Out])
	outCh := make(chan chan Out, workers*2) // The buffer helps to not block the workers, if one of the workers is slow, or if the code reading the output iterator is slow.
	go iterOrderedProducer(ctx, in, inCh, outCh)
	iterOrderedWorkers(ctx, workers, f, inCh, outCh)
	iterOrderedConsumer(cancel, outCh, yield)
}

// iterOrderedProducer reads values from the input iterator, sends them to the workers and the consumer.
func iterOrderedProducer[In, Out any](ctx context.Context, in iter.Seq[In], inCh chan<- iterOrderedValue[In, Out], outCh chan<- chan Out) {
	defer close(inCh) // Notify the workers that there are no more values to process.
	chPool := getChannelPool[Out]()
	for inV := range in {
		ch := chPool.Get()
		inCh <- iterOrderedValue[In, Out]{
			value: inV,
			ch:    ch,
		}
		outCh <- ch
		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

// iterOrderedWorkers starts the worker goroutines, and closes the output channel when all workers are done.
func iterOrderedWorkers[In, Out any](ctx context.Context, workers int, f func(context.Context, In) Out, inCh <-chan iterOrderedValue[In, Out], outCh chan<- chan Out) {
	wg := new(sync.WaitGroup)
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			iterOrderedWorker(ctx, f, inCh)
		}()
	}
	go func() {
		wg.Wait()
		close(outCh) // Notify the consumer that all workers are done.
	}()
}

// iterOrderedWorker reads the input value from the input channel, runs the function, and sends the output value to the output channel.
func iterOrderedWorker[In, Out any](ctx context.Context, f func(context.Context, In) Out, inCh <-chan iterOrderedValue[In, Out]) {
	for inV := range inCh {
		func() {
			defer panichandle.Recover(ctx)
			ok := false
			defer func() {
				if !ok {
					close(inV.ch) // Close the channel if the worker didn't send a value (panic).
				}
			}()
			outV := f(ctx, inV.value)
			inV.ch <- outV
			ok = true
		}()
	}
}

// iterOrderedConsumer reads the output values from the output channel and yields them to the output iterator.
func iterOrderedConsumer[Out any](cancel context.CancelFunc, outCh <-chan chan Out, yield func(Out) bool) {
	stopped := false
	chPool := getChannelPool[Out]()
	for ch := range outCh {
		outV, ok := <-ch
		if !ok {
			continue // Skip the value if the worker didn't send a value.
		}
		chPool.Put(ch)
		if stopped {
			continue // Drain the output channel if the output iterator was stopped.
		}
		if !yield(outV) {
			cancel() // Notify the producer that the output iterator was stopped.
			stopped = true
		}
	}
}

type iterOrderedValue[In, Out any] struct {
	value In
	ch    chan<- Out
}

var channelPools = syncutil.Map[reflect.Type, any]{}

func getChannelPool[T any]() *syncutil.Pool[chan T] {
	typ := reflect.TypeFor[T]()
	poolItf, _ := channelPools.Load(typ)
	pool, ok := poolItf.(*syncutil.Pool[chan T])
	if !ok {
		pool = &syncutil.Pool[chan T]{
			New: func() chan T {
				return make(chan T, 1) // The buffer helps to not block the workers if the consumer is not yet reading this value.
			},
		}
		channelPools.Store(typ, pool)
	}
	return pool
}

// Slice is a [Iter] wrapper for slices.
func Slice[SIn ~[]In, SOut []Out, In, Out any](ctx context.Context, in SIn, workers int, f func(ctx context.Context, i int, v In) Out) SOut {
	res := Iter(
		ctx,
		iterutil.Seq2ToSeq(slices.All(in), iterutil.NewKeyVal),
		min(workers, len(in)),
		func(ctx context.Context, kv iterutil.KeyVal[int, In]) iterutil.KeyVal[int, Out] {
			return iterutil.KeyVal[int, Out]{
				Key: kv.Key,
				Val: f(ctx, kv.Key, kv.Val),
			}
		},
	)
	out := make(SOut, len(in))
	for v := range res {
		out[v.Key] = v.Val
	}
	return out
}

// SliceError is a [Slice] wrapper that returns an error.
//
//nolint:dupl // This is not exactly the same code.
func SliceError[SIn ~[]In, SOut []Out, In, Out any](ctx context.Context, in SIn, workers int, f func(ctx context.Context, i int, v In) (Out, error)) (SOut, error) {
	res := Iter(
		ctx,
		iterutil.Seq2ToSeq(slices.All(in), iterutil.NewKeyVal),
		min(workers, len(in)),
		WithError(func(ctx context.Context, kv iterutil.KeyVal[int, In]) (iterutil.KeyVal[int, Out], error) {
			v, err := f(ctx, kv.Key, kv.Val)
			return iterutil.KeyVal[int, Out]{
				Key: kv.Key,
				Val: v,
			}, err
		}),
	)
	out := make(SOut, len(in))
	var errs []error
	for v := range res {
		out[v.Val.Key] = v.Val.Val
		if v.Err != nil {
			errs = append(errs, v.Err)
		}
	}
	err := errors.Join(errs...)
	return out, err
}

// Map is a [Iter] wrapper for maps.
func Map[MIn ~map[K]In, MOut map[K]Out, K comparable, In, Out any](ctx context.Context, in MIn, workers int, f func(ctx context.Context, k K, v In) Out) MOut {
	res := Iter(
		ctx,
		iterutil.Seq2ToSeq(maps.All(in), iterutil.NewKeyVal),
		min(workers, len(in)),
		func(ctx context.Context, kv iterutil.KeyVal[K, In]) iterutil.KeyVal[K, Out] {
			return iterutil.KeyVal[K, Out]{
				Key: kv.Key,
				Val: f(ctx, kv.Key, kv.Val),
			}
		},
	)
	out := make(MOut, len(in))
	for v := range res {
		out[v.Key] = v.Val
	}
	return out
}

// MapError is a [Map] wrapper that returns an error.
//
//nolint:dupl // This is not exactly the same code.
func MapError[MIn ~map[K]In, MOut map[K]Out, K comparable, In, Out any](ctx context.Context, in MIn, workers int, f func(ctx context.Context, k K, v In) (Out, error)) (MOut, error) {
	res := Iter(
		ctx,
		iterutil.Seq2ToSeq(maps.All(in), iterutil.NewKeyVal),
		min(workers, len(in)),
		WithError(func(ctx context.Context, kv iterutil.KeyVal[K, In]) (iterutil.KeyVal[K, Out], error) {
			v, err := f(ctx, kv.Key, kv.Val)
			return iterutil.KeyVal[K, Out]{
				Key: kv.Key,
				Val: v,
			}, err
		}),
	)
	out := make(MOut, len(in))
	var errs []error
	for v := range res {
		out[v.Val.Key] = v.Val.Val
		if v.Err != nil {
			errs = append(errs, v.Err)
		}
	}
	err := errors.Join(errs...)
	return out, err
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
