// Package panicutil provides utilities for panics.
package panicutil

import (
	"fmt"

	"github.com/pierrre/go-libs/bufpool"
	"github.com/pierrre/go-libs/runtimeutil"
)

// Error represents a panic error, which includes the recovered panic value and the stack trace at the time of the panic.
type Error struct {
	Recovered any
	Callers   []uintptr
}

// NewError creates a new error from a recovered panic value.
// The error message includes the panic value and the stack trace.
func NewError(r any) error {
	return &Error{
		Recovered: r,
		Callers:   runtimeutil.GetCallers(1),
	}
}

// Error implements error.
func (p *Error) Error() string {
	buf := bufPool.Get()
	defer bufPool.Put(buf)
	_, _ = fmt.Fprint(buf, p.Recovered)
	_, _ = buf.WriteString("\n\n")
	fs := runtimeutil.GetCallersFrames(p.Callers)
	_, _ = runtimeutil.WriteFrames(buf, fs)
	return buf.String()
}

// Unwrap returns the wrapped error if the recovered panic value is an error.
func (p *Error) Unwrap() error {
	err, _ := p.Recovered.(error)
	return err
}

// StackFrames returns the stack frames of the panic.
//
// It is used by the Sentry library.
func (p *Error) StackFrames() []uintptr {
	return p.Callers
}

var bufPool = &bufpool.Pool{
	MaxCap: -1,
}
