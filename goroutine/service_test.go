package goroutine

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/pierrre/assert"
)

func ExampleServices() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	fmt.Println("start")
	err := Services(ctx, map[string]func(context.Context) error{
		"a": func(_ context.Context) error { //nolint:unparam // It's a test.
			fmt.Println("service A")
			cancel()
			return nil
		},
		"b": func(ctx context.Context) error { //nolint:unparam // It's a test.
			<-ctx.Done()
			fmt.Println("service B")
			return nil
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("stop")
	// Output:
	// start
	// service A
	// service B
	// stop
}

func TestServices(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	err := Services(ctx, map[string]func(context.Context) error{
		"a": func(_ context.Context) error { //nolint:unparam // It's a test.
			cancel()
			return nil
		},
		"b": func(ctx context.Context) error { //nolint:unparam // It's a test.
			<-ctx.Done()
			return nil
		},
	})
	assert.NoError(t, err)
}

func TestServicesError(t *testing.T) {
	ctx := context.Background()
	err := Services(ctx, map[string]func(context.Context) error{
		"a": func(_ context.Context) error {
			return errors.New("error")
		},
		"b": func(ctx context.Context) error { //nolint:unparam // It's a test.
			<-ctx.Done()
			return nil
		},
	})
	assert.Error(t, err)
}
