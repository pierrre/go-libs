package goroutine

import (
	"context"
	"iter"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/pierrre/go-libs/iterutil"
	"github.com/pierrre/go-libs/syncutil"
)

// Iter runs a function for each values of an input [iter.Seq] with concurrent workers.
// It returns an [iter.Seq] of unordered output values.
// For an ordered version, see [IterOrdered].
//
// If the [context.Context] is canceled, it stops reading values from the input.
// If the caller stops iterating the output, the [context.Context] is canceled.
func Iter[In, Out any](ctx context.Context, in iter.Seq[In], workers int, f func(context.Context, In) Out) iter.Seq[Out] {
	workers = max(workers, 1) // We need at least 1 worker.
	return func(yield func(Out) bool) {
		ctx, cancel := context.WithCancel(ctx) //nolint:govet // Shadowing is expected here.
		defer cancel()
		inCh := make(chan In)
		outCh := make(chan Out, workers)             // The buffer prevents blocking the workers if the output iterator is slow.
		defer Start(ctx, func(ctx context.Context) { // Send values from the input iterator to the workers.
			defer close(inCh)      // Notify the workers that there are no more value.
			in(func(inV In) bool { // Read values from the input iterator.
				inCh <- inV             // Send value to the workers.
				return ctx.Err() == nil // Stop sending values if the context is canceled.
			})
		}).Wait() // Wait until the producer is stopped.
		runningWorkers := int64(workers)                              // Count of running workers.
		defer startN(ctx, workers, func(ctx context.Context, _ int) { // Start the workers.
			defer func() { // When the workers are stopped.
				if atomic.AddInt64(&runningWorkers, -1) == 0 { // Wait for all workers to finish.
					close(outCh)            // Notify the consumer that there are no more value.
					drainChannel(inCh, nil) // Consume remaining values to avoid blocking the producer.
				}
			}()
			for inV := range inCh { // Read values from the producer.
				outCh <- f(ctx, inV) // Process value with the function and send the result to the consumer.
			}
		}).Wait() // Wait until the workers are stopped.
		defer func() { // When the output iterator is stopped.
			cancel()                 // Notify the producer to stop sending values.
			drainChannel(outCh, nil) // Consume remaining values to avoid blocking the workers.
		}()
		for outV := range outCh { // Consume values from the workers.
			if !yield(outV) { // Send value to the output iterator.
				return
			}
		}
	}
}

// IterOrdered is like [Iter] but it preserves the order of values.
func IterOrdered[In, Out any](ctx context.Context, in iter.Seq[In], workers int, f func(context.Context, In) Out) iter.Seq[Out] {
	workers = max(workers, 1)                  // We need at least 1 worker.
	pool := getIterOrderedValuePool[In, Out]() // Recycle values to avoid allocations.
	return func(yield func(Out) bool) {
		ctx, cancel := context.WithCancel(ctx) //nolint:govet // Shadowing is expected here.
		defer cancel()
		inCh := make(chan *iterOrderedValue[In, Out])
		outCh := make(chan *iterOrderedValue[In, Out], workers*2) // The buffer prevents blocking the workers if one of the workers or the output iterator is slow.
		defer Start(ctx, func(ctx context.Context) {              // Send values from the input iterator to the workers annd the consumer.
			defer func() {
				close(inCh)  // Notify the workers that there are no more value.
				close(outCh) // Notify the consumer that there are no more value.
			}()
			in(func(inV In) bool { // Read values from the input iterator.
				v := pool.Get()         // Get a value from the pool.
				v.wg.Add(1)             // Enforce the consumer to wait for the worker to finish processing the value.
				v.in = inV              // Set the input value.
				inCh <- v               // Send value to the workers.
				outCh <- v              // Send value to the consumer.
				return ctx.Err() == nil // Stop sending values if the context is canceled.
			})
		}).Wait() // Wait until the producer is stopped.
		runningWorkers := int64(workers)
		defer startN(ctx, workers, func(ctx context.Context, _ int) { // Start the workers.
			defer func() { // When the workers are stopped.
				cancel()                                       // Notify the producer to stop sending values (required to handle panic and [runtime.Goexit]).
				if atomic.AddInt64(&runningWorkers, -1) == 0 { // Wait for all workers to finish.
					drainChannel(inCh, func(v *iterOrderedValue[In, Out]) { // Consume remaining values to avoid blocking the producer.
						v.wg.Done() // Notify the consumer that the worker is done processing the value (but it didn't process it).
					})
				}
			}()
			for v := range inCh { // Read values from the producer.
				func() {
					defer v.wg.Done()    // Notify the consumer that the worker is done processing the value.
					v.out = f(ctx, v.in) // Process value with the function and set the output value.
					v.ok = true          // Notify the consumer that the worker processed the value successfully.
				}()
			}
		}).Wait() // Wait until the workers are stopped.
		defer func() { // When the output iterator is stopped.
			cancel()                                                 // Notify the producer to stop sending values.
			drainChannel(outCh, func(v *iterOrderedValue[In, Out]) { // Consume remaining values to avoid blocking the workers.
				v.wg.Wait()     // Wait for the worker to finish processing the value.
				v.release(pool) // Release the value back to the pool.
			})
		}()
		for v := range outCh { // Consume values from the workers.
			v.wg.Wait()             // Wait for the worker to finish processing the value.
			outV, ok := v.out, v.ok // Store the result in local variables.
			v.release(pool)         // Release the value back to the pool.
			if !ok {                // If the worker didn't process the value successfully, skip it (panic or [runtime.Goexit]).
				continue
			}
			if !yield(outV) { // Send value to the output iterator.
				return
			}
		}
	}
}

type iterOrderedValue[In, Out any] struct {
	wg  sync.WaitGroup
	in  In
	out Out
	ok  bool
}

func (v *iterOrderedValue[In, Out]) release(pool *syncutil.Pool[*iterOrderedValue[In, Out]]) {
	*v = iterOrderedValue[In, Out]{}
	pool.Put(v)
}

type iterOrderedValuePoolsKey struct {
	in  reflect.Type
	out reflect.Type
}

var iterOrderedValuePools syncutil.Map[iterOrderedValuePoolsKey, any]

func getIterOrderedValuePool[In, Out any]() *syncutil.Pool[*iterOrderedValue[In, Out]] {
	key := iterOrderedValuePoolsKey{
		in:  reflect.TypeFor[In](),
		out: reflect.TypeFor[Out](),
	}
	poolItf, _ := iterOrderedValuePools.Load(key)
	pool, ok := poolItf.(*syncutil.Pool[*iterOrderedValue[In, Out]])
	if ok {
		return pool
	}
	pool = &syncutil.Pool[*iterOrderedValue[In, Out]]{
		New: func() *iterOrderedValue[In, Out] {
			return new(iterOrderedValue[In, Out])
		},
	}
	poolItf = pool
	poolItf, _ = iterOrderedValuePools.LoadOrStore(key, poolItf)
	pool, _ = poolItf.(*syncutil.Pool[*iterOrderedValue[In, Out]])
	return pool
}

func drainChannel[T any](ch <-chan T, f func(v T)) {
	for v := range ch {
		if f != nil {
			f(v)
		}
	}
}

// Iter2 is like [Iter] for [iter.Seq2].
func Iter2[K, In, Out any](ctx context.Context, in iter.Seq2[K, In], workers int, f func(context.Context, iterutil.KeyVal[K, In]) Out) iter.Seq2[K, Out] {
	return iter2(ctx, in, workers, f, Iter)
}

// Iter2Ordered is like [IterOrdered] for [iter.Seq2].
func Iter2Ordered[K, In, Out any](ctx context.Context, in iter.Seq2[K, In], workers int, f func(context.Context, iterutil.KeyVal[K, In]) Out) iter.Seq2[K, Out] {
	return iter2(ctx, in, workers, f, IterOrdered)
}

type iterFunc[In, Out any] func(ctx context.Context, in iter.Seq[In], workers int, f func(context.Context, In) Out) iter.Seq[Out]

func iter2[K, In, Out any](ctx context.Context, in iter.Seq2[K, In], workers int, f func(context.Context, iterutil.KeyVal[K, In]) Out, ifn iterFunc[iterutil.KeyVal[K, In], iterutil.KeyVal[K, Out]]) iter.Seq2[K, Out] {
	return iterutil.SeqToSeq2(
		ifn(
			ctx,
			iterutil.Seq2ToSeq(in, iterutil.NewKeyVal),
			workers,
			func(ctx context.Context, kv iterutil.KeyVal[K, In]) iterutil.KeyVal[K, Out] {
				return iterutil.KeyVal[K, Out]{
					Key: kv.Key,
					Val: f(ctx, kv),
				}
			},
		),
		iterutil.KeyVal[K, Out].Values,
	)
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
