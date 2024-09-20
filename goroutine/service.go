package goroutine

import (
	"context"
	"errors"
	"fmt"

	"github.com/pierrre/go-libs/panichandle"
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
	errCh := make(chan error)
	wg := waitGroupPool.Get()
	wg.Add(len(services))
	for name, service := range services {
		go func() {
			defer panichandle.Recover(ctx)
			defer wg.Done()
			err := service(ctx)
			if err != nil {
				err = fmt.Errorf("%s: %w", name, err)
				errCh <- err
			}
		}()
	}
	go func() {
		wg.Wait()
		waitGroupPool.Put(wg)
		close(errCh)
	}()
	var errs []error //nolint:prealloc // We don't know the number of errors.
	for err := range errCh {
		cancel()
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}
