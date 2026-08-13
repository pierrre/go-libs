package goroutine

import (
	"context"
	"errors"
	"slices"

	"github.com/pierrre/go-libs/iterutil"
)

// Slice processes a slice with [Iter2].
// The output slice has the same length as the input slice, and each output element corresponds to the input element at the same index.
// The workers parameter is capped to the length of the input slice.
func Slice[SIn ~[]In, SOut []Out, In, Out any](ctx context.Context, in SIn, workers int, f func(ctx context.Context, i int, v In) Out) SOut {
	var out SOut
	if in != nil {
		out = make(SOut, len(in))
		if len(in) > 0 {
			res := Iter2(ctx, slices.All(in), min(workers, len(in)), func(ctx context.Context, iv iterutil.KeyVal[int, In]) Out {
				return f(ctx, iv.Key, iv.Val)
			})
			res(func(i int, v Out) bool {
				out[i] = v
				return true
			})
		}
	}
	return out
}

// SliceError is like [Slice] but returns an error.
// The output slice has the same length as the input slice, and each output element corresponds to the input element at the same index.
// The workers parameter is capped to the length of the input slice.
func SliceError[SIn ~[]In, SOut []Out, In, Out any](ctx context.Context, in SIn, workers int, f func(ctx context.Context, i int, v In) (Out, error)) (SOut, error) {
	var out SOut
	var errs []error
	if in != nil {
		out = make(SOut, len(in))
		if len(in) > 0 {
			res := Iter2(ctx, slices.All(in), min(workers, len(in)), WithError(func(ctx context.Context, iv iterutil.KeyVal[int, In]) (Out, error) {
				return f(ctx, iv.Key, iv.Val)
			}))
			res(func(i int, ve ValErr[Out]) bool {
				out[i] = ve.Val
				if ve.Err != nil {
					errs = append(errs, ve.Err)
				}
				return true
			})
		}
	}
	return out, errors.Join(errs...)
}

// SliceFunc processes a slice of functions.
func SliceFunc[SOut []Out, Out any](ctx context.Context, fs []func(ctx context.Context) Out, workers int) SOut {
	return Slice(ctx, fs, workers, func(ctx context.Context, i int, f func(ctx context.Context) Out) Out {
		return f(ctx)
	})
}

// SliceFuncError processes a slice of functions that return an error.
func SliceFuncError[SOut []Out, Out any](ctx context.Context, fs []func(ctx context.Context) (Out, error), workers int) (SOut, error) {
	return SliceError(ctx, fs, workers, func(ctx context.Context, i int, f func(ctx context.Context) (Out, error)) (Out, error) {
		return f(ctx)
	})
}
