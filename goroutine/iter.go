package goroutine

import (
	"context"
	"iter"
	"reflect"

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

func iterSeq[In, Out any](ctx context.Context, in iter.Seq[In], workers int, f func(context.Context, In) Out, yield func(Out) bool) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	workers = max(workers, 1)
	inCh := make(chan In)
	outCh := make(chan Out, workers)
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

// IterOrdered is like [Iter] but it keeps the order of the output values.
func IterOrdered[In, Out any](ctx context.Context, in iter.Seq[In], workers int, f func(context.Context, In) Out) iter.Seq[Out] {
	return func(yield func(Out) bool) {
		iterOrdered(ctx, in, workers, f, yield)
	}
}

func iterOrdered[In, Out any](ctx context.Context, in iter.Seq[In], workers int, f func(context.Context, In) Out, yield func(Out) bool) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	workers = max(workers, 1)
	inCh := make(chan iterOrderedValue[In, Out])
	outCh := make(chan chan Out, workers*2)
	go iterOrderedProducer(ctx, in, inCh, outCh)
	iterOrderedWorkers(ctx, workers, f, inCh, outCh)
	iterOrderedConsumer(cancel, outCh, yield)
}

func iterOrderedProducer[In, Out any](ctx context.Context, in iter.Seq[In], inCh chan<- iterOrderedValue[In, Out], outCh chan<- chan Out) {
	defer close(inCh)
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

func iterOrderedWorkers[In, Out any](ctx context.Context, workers int, f func(context.Context, In) Out, inCh <-chan iterOrderedValue[In, Out], outCh chan<- chan Out) {
	wg := waitGroupPool.Get()
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			iterOrderedWorker(ctx, f, inCh)
		}()
	}
	go func() {
		wg.Wait()
		waitGroupPool.Put(wg)
		close(outCh)
	}()
}

func iterOrderedWorker[In, Out any](ctx context.Context, f func(context.Context, In) Out, inCh <-chan iterOrderedValue[In, Out]) {
	for inV := range inCh {
		func() {
			defer panichandle.Recover(ctx)
			ok := false
			defer func() {
				if !ok {
					close(inV.ch)
				}
			}()
			outV := f(ctx, inV.value)
			ok = true
			inV.ch <- outV
		}()
	}
}

func iterOrderedConsumer[Out any](cancel context.CancelFunc, outCh <-chan chan Out, yield func(Out) bool) {
	stopped := false
	chPool := getChannelPool[Out]()
	for ch := range outCh {
		outV, ok := <-ch
		if !ok {
			continue
		}
		chPool.Put(ch)
		if stopped {
			continue
		}
		if !yield(outV) {
			cancel()
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
				return make(chan T, 1)
			},
		}
		channelPools.Store(typ, pool)
	}
	return pool
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
