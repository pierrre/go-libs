package goroutine

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/pierrre/go-libs/iterutil"
)

// Services runs multiple services in goroutines.
//
// It waits for all services to finish.
//
// If a service returns an error, it cancels the context.
// All errors are joined and returned.
func Services(ctx context.Context, services map[string]func(context.Context) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errs := Iter(ctx, iterutil.Seq2ToSeq(maps.All(services), iterutil.NewKeyVal), len(services), func(ctx context.Context, service iterutil.KeyVal[string, func(context.Context) error]) error {
		err := service.Val(ctx)
		if err != nil {
			cancel()
			return fmt.Errorf("%s: %w", service.Key, err)
		}
		return nil
	})
	return errors.Join(slices.Collect(errs)...)
}
