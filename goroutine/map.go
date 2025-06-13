//nolint:dupl // Similar code but different types.
package goroutine

import (
	"context"
	"errors"
	"maps"

	"github.com/pierrre/go-libs/iterutil"
)

// Map processes a map with [Iter].
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
