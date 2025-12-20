//nolint:dupl // Similar code but different types.
package goroutine

import (
	"context"
	"errors"
	"slices"
)

// Slice processes a slice with [Iter].
func Slice[SIn ~[]In, SOut []Out, In, Out any](ctx context.Context, in SIn, workers int, f func(ctx context.Context, i int, v In) Out) SOut {
	res := iterCollection(ctx, slices.All(in), min(workers, len(in)), f)
	out := make(SOut, len(in))
	for v := range res {
		out[v.Key] = v.Val
	}
	return out
}

// SliceError is a [Slice] wrapper that returns an error.
func SliceError[SIn ~[]In, SOut []Out, In, Out any](ctx context.Context, in SIn, workers int, f func(ctx context.Context, i int, v In) (Out, error)) (SOut, error) {
	res := iterCollectionError(ctx, slices.All(in), min(workers, len(in)), f)
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
