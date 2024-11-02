// Package ctxstack provides a way to store stack traces in a context.
package ctxstack

import (
	"context"
	"iter"

	"github.com/pierrre/go-libs/runtimeutil"
)

type contextKey struct{}

type stackCtx struct {
	context.Context //nolint:containedctx // stackCtx is a custom context type and not a value, because it helps to reduce memory allocations.
	stack           []uintptr
	next            *stackCtx
}

func (ctx *stackCtx) Value(key any) any {
	if key == (contextKey{}) {
		return ctx
	}
	return ctx.Context.Value(key)
}

// NewContext creates a new [context.Context] with the current stack trace.
func NewContext(ctx context.Context) context.Context {
	return &stackCtx{
		Context: ctx,
		stack:   runtimeutil.GetCallers(1),
		next:    fromContext(ctx),
	}
}

// FromContext returns a [iter.Seq] of stack traces stored in the [context.Context].
func FromContext(ctx context.Context) iter.Seq[[]uintptr] {
	return func(yield func([]uintptr) bool) {
		for v := fromContext(ctx); v != nil; v = v.next {
			if !yield(v.stack) {
				return
			}
		}
	}
}

func fromContext(ctx context.Context) *stackCtx {
	v, _ := ctx.Value(contextKey{}).(*stackCtx)
	return v
}
