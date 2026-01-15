package goroutine

import (
	"context"
	"errors"
	"slices"

	"github.com/pierrre/go-libs/iterutil"
)

// Slice processes a slice with [Iter].
func Slice[SIn ~[]In, SOut []Out, In, Out any](ctx context.Context, in SIn, workers int, f func(ctx context.Context, i int, v In) Out) SOut {
	res := Iter2(ctx, slices.All(in), min(workers, len(in)), func(ctx context.Context, iv iterutil.KeyVal[int, In]) Out {
		return f(ctx, iv.Key, iv.Val)
	})
	out := make(SOut, len(in))
	for i, v := range res {
		out[i] = v
	}
	return out
}

// SliceError is a [Slice] wrapper that returns an error.
func SliceError[SIn ~[]In, SOut []Out, In, Out any](ctx context.Context, in SIn, workers int, f func(ctx context.Context, i int, v In) (Out, error)) (SOut, error) {
	res := Iter2(ctx, slices.All(in), min(workers, len(in)), WithError(func(ctx context.Context, iv iterutil.KeyVal[int, In]) (Out, error) {
		return f(ctx, iv.Key, iv.Val)
	}))
	out := make(SOut, len(in))
	var errs []error
	for i, ve := range res {
		out[i] = ve.Val
		if ve.Err != nil {
			errs = append(errs, ve.Err)
		}
	}
	err := errors.Join(errs...)
	return out, err
}
