// Package panicutil provides utilities for panics.
package panicutil

import (
	"fmt"

	"github.com/pierrre/go-libs/bufpool"
	"github.com/pierrre/go-libs/runtimeutil"
)

type panicError struct {
	r       any
	callers []uintptr
}

// NewError creates a new error from a recovered panic value.
// The error message includes the panic value and the stack trace.
func NewError(r any) error {
	return &panicError{
		r:       r,
		callers: runtimeutil.GetCallers(1),
	}
}

// Error implements error.
func (p *panicError) Error() string {
	buf := bufPool.Get()
	defer bufPool.Put(buf)
	_, _ = fmt.Fprint(buf, p.r)
	_, _ = buf.WriteString("\n\n")
	fs := runtimeutil.GetCallersFrames(p.callers)
	_, _ = runtimeutil.WriteFrames(buf, fs)
	return buf.String()
}

// Unwrap returns the wrapped error if the recovered panic value is an error.
func (p *panicError) Unwrap() error {
	err, _ := p.r.(error)
	return err
}

func (p *panicError) StackFrames() []uintptr {
	return p.callers
}

var bufPool = &bufpool.Pool{
	MaxCap: -1,
}
