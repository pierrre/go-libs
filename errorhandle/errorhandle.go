// Package errorhandle handles error.
package errorhandle

import (
	"context"
	"fmt"
	"os"
)

// DefaultHandler is the default [Handler].
var DefaultHandler Handler = StderrHandler

// Handler is a function that handles error.
type Handler func(ctx context.Context, err error)

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

// Handlers is a list of [Handler].
//
// They are called in order.
type Handlers []Handler

func (hs Handlers) Handle(ctx context.Context, err error) {
	for _, h := range hs {
		h(ctx, err)
	}
}

// FilterHandler is a [Handler] that filters error.
//
// If Filter returns true, the error is passed to Handler.
type FilterHandler struct {
	Handler
	Filter func(ctx context.Context, err error) bool
}

func (f FilterHandler) Handle(ctx context.Context, err error) {
	if f.Filter(ctx, err) {
		f.Handler(ctx, err)
	}
}

// StderrHandler is a [Handler] that writes the error to os.Stderr.
func StderrHandler(ctx context.Context, err error) {
	_, _ = fmt.Fprintln(os.Stderr, err)
}
