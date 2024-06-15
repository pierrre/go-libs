// Package errorhandle handles error.
package errorhandle

import (
	"context"
	"fmt"
	"os"
)

// Handle handles error.
var Handle Handler = StderrHandler

// Handler is a function that handles error.
type Handler func(ctx context.Context, err error)

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
