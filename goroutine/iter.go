package goroutine

import (
	"context"
	"iter"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/pierrre/go-libs/syncutil"
)

// Iter runs a function for each values of an input iterator with concurrent workers.
// It returns an iterator with the unordered output values.
// For an ordered version, see [IterOrdered].
//
// If the context is canceled, it stops reading values from the input iterator.
// If the caller stops iterating the output iterator, the context is canceled.
func Iter[In, Out any](parentCtx context.Context, in iter.Seq[In], workers int, f func(context.Context, In) Out) iter.Seq[Out] {
	workers = max(workers, 1)
	return func(yield func(Out) bool) {
		ctx, cancel := context.WithCancel(parentCtx)
		defer cancel()
		inCh := make(chan In)
		outCh := make(chan Out, workers) // Prevent blocking the workers if the output iterator is slow.
		defer Start(ctx, func(ctx context.Context) {
			defer close(inCh)
			in(func(inV In) bool {
				inCh <- inV
				return ctx.Err() == nil
			})
		}).Wait()
		runningWorkers := int64(workers)
		defer startN(ctx, workers, func(ctx context.Context) {
			defer func() {
				if atomic.AddInt64(&runningWorkers, -1) == 0 {
					close(outCh)
					drainChannel(inCh, nil)
				}
			}()
			for inV := range inCh {
				outCh <- f(ctx, inV)
			}
		}).Wait()
		defer func() {
			cancel()
			drainChannel(outCh, nil)
		}()
		for outV := range outCh {
			if !yield(outV) {
				return
			}
		}
	}
}

// IterOrdered is like [Iter] but it keeps the order of the output values.
func IterOrdered[In, Out any](parentCtx context.Context, in iter.Seq[In], workers int, f func(context.Context, In) Out) iter.Seq[Out] {
	workers = max(workers, 1)
	pool := getIterOrderedValuePool[In, Out]()
	return func(yield func(Out) bool) {
		ctx, cancel := context.WithCancel(parentCtx)
		defer cancel()
		inCh := make(chan *iterOrderedValue[In, Out])
		outCh := make(chan *iterOrderedValue[In, Out], workers*2) // Prevent blocking the workers if one of the workers or the output iterator is slow.
		defer Start(ctx, func(ctx context.Context) {
			defer func() {
				close(inCh)
				close(outCh)
			}()
			in(func(inV In) bool {
				v := pool.Get()
				v.wg.Add(1)
				v.in = inV
				inCh <- v
				outCh <- v
				return ctx.Err() == nil
			})
		}).Wait()
		runningWorkers := int64(workers)
		defer startN(ctx, workers, func(ctx context.Context) {
			defer func() {
				cancel()
				if atomic.AddInt64(&runningWorkers, -1) == 0 {
					drainChannel(inCh, func(v *iterOrderedValue[In, Out]) {
						v.wg.Done()
					})
				}
			}()
			for v := range inCh {
				func() {
					defer v.wg.Done()
					v.out = f(ctx, v.in)
					v.ok = true
				}()
			}
		}).Wait()
		defer func() {
			cancel()
			drainChannel(outCh, func(v *iterOrderedValue[In, Out]) {
				v.wg.Wait()
				v.release(pool)
			})
		}()
		for v := range outCh {
			v.wg.Wait()
			outV, ok := v.out, v.ok
			v.release(pool)
			if !ok {
				continue
			}
			if !yield(outV) {
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
