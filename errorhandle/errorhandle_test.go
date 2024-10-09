package errorhandle

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/pierrre/assert"
)

func Example() {
	ctx := context.Background()
	ctx = SetHandlerToContext(ctx, func(ctx context.Context, err error) {
		fmt.Println("Error:", err)
	})
	Handle(ctx, errors.New("test"))
	// Output: Error: test
}

func TestSetHandlerToContext(t *testing.T) {
	ctx := context.Background()
	h := func(ctx context.Context, err error) {}
	ctx = SetHandlerToContext(ctx, h)
	h = GetHandlerFromContext(ctx)
	assert.True(t, h != nil)
}

func TestGetHandlerFromContextNotSet(t *testing.T) {
	ctx := context.Background()
	h := GetHandlerFromContext(ctx)
	assert.True(t, h == nil)
}

func TestGetHandlerContextSet(t *testing.T) {
	ctx := context.Background()
	h := func(ctx context.Context, err error) {}
	ctx = SetHandlerToContext(ctx, h)
	h = GetHandler(ctx)
	assert.True(t, h != nil)
}

func TestGetHandlerDefault(t *testing.T) {
	ctx := context.Background()
	h := GetHandler(ctx)
	assert.True(t, h != nil)
}

func TestHandle(t *testing.T) {
	ctx := context.Background()
	called := false
	h := func(ctx context.Context, err error) {
		called = true
	}
	ctx = SetHandlerToContext(ctx, h)
	Handle(ctx, errors.New("error"))
	assert.True(t, called)
}

func TestHandlers(t *testing.T) {
	ctx := context.Background()
	called := make([]int, 2)
	hs := Handlers{
		func(ctx context.Context, err error) {
			assert.DeepEqual(t, []int{0, 0}, called)
			called[0]++
		},
		func(ctx context.Context, err error) {
			assert.DeepEqual(t, []int{1, 0}, called)
			called[1]++
		},
	}
	hs.Handle(ctx, errors.New("error"))
	assert.DeepEqual(t, []int{1, 1}, called)
}

func TestFilterHandlerTrue(t *testing.T) {
	ctx := context.Background()
	called := false
	fh := FilterHandler{
		Handler: func(ctx context.Context, err error) {
			called = true
		},
		Filter: func(ctx context.Context, err error) bool {
			return true
		},
	}
	fh.Handle(ctx, errors.New("error"))
	assert.True(t, called)
}

func TestFilterHandlerFalse(t *testing.T) {
	ctx := context.Background()
	fh := FilterHandler{
		Handler: func(ctx context.Context, err error) {
			t.Fatal("should not be called")
		},
		Filter: func(ctx context.Context, err error) bool {
			return false
		},
	}
	fh.Handle(ctx, errors.New("error"))
}
