//nolint:dupl // Similar code but different types.
package goroutine

import (
	"context"
	"errors"
	"slices"

	"github.com/pierrre/go-libs/iterutil"
)

// Slice processes a slice with [Iter].
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
