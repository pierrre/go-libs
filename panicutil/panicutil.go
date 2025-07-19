// Package panicutil provides utilities for panics.
package panicutil

import (
	"bytes"
	"fmt"
	"runtime/debug"
)

type panicError struct {
	r     any
	stack []byte
}

// NewError creates a new error from a recovered panic value.
// The error message includes the panic value and the stack trace.
func NewError(r any) error {
	stack := debug.Stack()
	// The first line of the stack trace is of the form "goroutine N [status]:"
	// but by the time the panic reaches Do the goroutine may no longer exist
	// and its status will have changed. Trim out the misleading line.
	if line := bytes.IndexByte(stack, '\n'); line >= 0 {
		stack = stack[line+1:]
	}
	return &panicError{r: r, stack: stack}
}

// Error implements error.
func (p *panicError) Error() string {
	return fmt.Sprintf("%v\n\n%s", p.r, p.stack)
}

// Unwrap returns the wrapped error if the recovered panic value is an error.
func (p *panicError) Unwrap() error {
	err, _ := p.r.(error)
	return err
}
