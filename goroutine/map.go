package goroutine

import (
	"context"
	"errors"
	"maps"

	"github.com/pierrre/go-libs/iterutil"
)

// Map processes a map with [Iter].
func Map[MIn ~map[K]In, MOut map[K]Out, K comparable, In, Out any](ctx context.Context, in MIn, workers int, f func(ctx context.Context, kv iterutil.KeyVal[K, In]) Out) MOut {
	res := Iter2(ctx, maps.All(in), min(workers, len(in)), f)
	out := make(MOut, len(in))
	maps.Insert(out, res)
	return out
}

// MapError is a [Map] wrapper that returns an error.
func MapError[MIn ~map[K]In, MOut map[K]Out, K comparable, In, Out any](ctx context.Context, in MIn, workers int, f func(ctx context.Context, kv iterutil.KeyVal[K, In]) (Out, error)) (MOut, error) {
	res := Iter2(ctx, maps.All(in), min(workers, len(in)), WithError(f))
	out := make(MOut, len(in))
	var errs []error
	for k, ve := range res {
		out[k] = ve.Val
		if ve.Err != nil {
			errs = append(errs, ve.Err)
		}
	}
	err := errors.Join(errs...)
	return out, err
}
