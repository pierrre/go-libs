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
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ne := Iter2(ctx, maps.All(services), len(services), func(ctx context.Context, service iterutil.KeyVal[string, func(context.Context) error]) error {
		return service.Val(ctx)
	})
	var errs []error
	for name, err := range ne {
		if err != nil {
			cancel()
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
		}
	}
	return errors.Join(errs...)
}
