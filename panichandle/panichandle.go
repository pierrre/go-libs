// Package panichandle handles panic.
package panichandle

import (
	"context"
)

// DefaultHandler is the default [Handler].
//
// By default there is no [Handler].
var DefaultHandler Handler

// Recover recovers panic and calls the [Handler] returned by [GetHandler].
//
// If there is no [Handler], it doesn't recover.
//
// It should be called in defer.
func Recover(ctx context.Context) {
	h := GetHandler(ctx)
	if h != nil {
		r := recover()
		if r != nil {
			h(ctx, r)
		}
	}
}

// Handler is a function that handles panic.
type Handler func(ctx context.Context, r any)

type contextKey struct{}

// SetHandlerToContext sets a [Handler] to a [context.Context].
func SetHandlerToContext(ctx context.Context, h Handler) context.Context {
	return context.WithValue(ctx, contextKey{}, h)
}

// GetHandlerFromContext gets a [Handler] from a [context.Context].
//
// It returns nil if no [Handler] is set.
func GetHandlerFromContext(ctx context.Context) Handler {
	h, _ := ctx.Value(contextKey{}).(Handler)
	return h
}

// GetHandler gets a [Handler] from a [context.Context] or the DefaultHandler.
func GetHandler(ctx context.Context) Handler {
	h := GetHandlerFromContext(ctx)
	if h != nil {
		return h
	}
	return DefaultHandler
}

// ErrorHandler is a [Handler] that converts the recovered value to [error] with Convert, and calls Handler.
//
// If the recovered value is already an [error], Convert is not called.
type ErrorHandler struct {
	Handler func(ctx context.Context, err error)
	Convert func(r any) error
}

func (h ErrorHandler) Handle(ctx context.Context, r any) {
	err, ok := r.(error)
	if !ok {
		err = h.Convert(r)
	}
	h.Handler(ctx, err)
}
