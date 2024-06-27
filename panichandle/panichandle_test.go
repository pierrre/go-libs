package panichandle

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/pierrre/assert"
)

func TestRecoverNoPanicWithoutHandler(t *testing.T) {
	ctx := context.Background()
	Recover(ctx)
}

func TestRecoverNoPanicWitHandler(t *testing.T) {
	ctx := context.Background()
	defer func() {
		DefaultHandler = nil
	}()
	called := false
	DefaultHandler = func(ctx context.Context, r any) {
		called = true
	}
	defer func() {
		assert.False(t, called)
	}()
	Recover(ctx)
}

func TestRecoverPanicWithoutHandler(t *testing.T) {
	ctx := context.Background()
	defer func() {
		r := recover()
		assert.NotZero(t, r)
	}()
	defer Recover(ctx)
	panic("test")
}

func TestRecoverPanicWithHandler(t *testing.T) {
	ctx := context.Background()
	defer func() {
		DefaultHandler = nil
	}()
	called := false
	DefaultHandler = func(ctx context.Context, r any) {
		called = true
		assert.NotZero(t, r)
	}
	defer func() {
		assert.True(t, called)
	}()
	defer Recover(ctx)
	panic("test")
}

func TestSetHandlerToContext(t *testing.T) {
	ctx := context.Background()
	h := func(ctx context.Context, r any) {}
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
	h := func(ctx context.Context, r any) {}
	ctx = SetHandlerToContext(ctx, h)
	h = GetHandler(ctx)
	assert.True(t, h != nil)
}

func TestGetHandlerDefault(t *testing.T) {
	ctx := context.Background()
	h := GetHandler(ctx)
	assert.True(t, h == nil)
}

func TestErrorHandlerString(t *testing.T) {
	ctx := context.Background()
	called := false
	h := ErrorHandler{
		Handler: func(ctx context.Context, err error) {
			called = true
			assert.ErrorEqual(t, err, "error")
		},
		Convert: func(r any) error {
			return fmt.Errorf("%s", r)
		},
	}
	h.Handle(ctx, "error")
	assert.True(t, called)
}

func TestErrorHandlerError(t *testing.T) {
	ctx := context.Background()
	called := false
	h := ErrorHandler{
		Handler: func(ctx context.Context, err error) {
			called = true
			assert.ErrorEqual(t, err, "error")
		},
	}
	h.Handle(ctx, errors.New("error"))
	assert.True(t, called)
}
