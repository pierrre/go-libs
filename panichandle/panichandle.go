// Package panichandle handles panic.
package panichandle

import (
	"context"
)

// Handle handles panic.
//
// By default there is no [Handler].
var Handle Handler

// Recover recovers panic and calls [Handle].
//
// If there is no [Handler], it doesn't recover.
//
// It should be called in defer.
func Recover(ctx context.Context) {
	if Handle != nil {
		r := recover()
		if r != nil {
			Handle(ctx, r)
		}
	}
}

// Handler is a function that handles panic.
type Handler func(ctx context.Context, r any)

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
