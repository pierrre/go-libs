package goroutine

import (
	"context"
	"errors"
	"fmt"
	"maps"

	"github.com/pierrre/go-libs/iterutil"
)

// Services runs multiple services in goroutines.
//
// It waits for all services to finish.
//
// If a service returns an error, it cancels the context.
// All errors are joined and returned.
func Services(ctx context.Context, services map[string]func(context.Context) error) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	ne := Iter2(ctx, maps.All(services), len(services), func(ctx context.Context, service iterutil.KeyVal[string, func(context.Context) error]) error {
		return service.Val(ctx)
	})
	var errs []error
	ne(func(name string, err error) bool {
		if err != nil {
			cancel(err)
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
		}
		return true
	})
	return errors.Join(errs...)
}
