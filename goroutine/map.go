//nolint:dupl // Similar code but different types.
package goroutine

import (
	"context"
	"errors"
	"maps"
)

// Map processes a map with [Iter].
func Map[MIn ~map[K]In, MOut map[K]Out, K comparable, In, Out any](ctx context.Context, in MIn, workers int, f func(ctx context.Context, k K, v In) Out) MOut {
	res := iterCollection(ctx, maps.All(in), min(workers, len(in)), f)
	out := make(MOut, len(in))
	for v := range res {
		out[v.Key] = v.Val
	}
	return out
}

// MapError is a [Map] wrapper that returns an error.
func MapError[MIn ~map[K]In, MOut map[K]Out, K comparable, In, Out any](ctx context.Context, in MIn, workers int, f func(ctx context.Context, k K, v In) (Out, error)) (MOut, error) {
	res := iterCollectionError(ctx, maps.All(in), min(workers, len(in)), f)
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
