// Package panicutil provides utilities for panics.
package panicutil

import (
	"fmt"

	"github.com/pierrre/go-libs/bytesutil"
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
	bw := bytesWriterPool.Get()
	defer bytesWriterPool.Put(bw)
	*bw = fmt.Append(*bw, p.Recovered)
	bw.AppendString("\n\n")
	*bw = runtimeutil.AppendCallersFrames(*bw, p.Callers)
	return bw.String()
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

var bytesWriterPool = &bytesutil.WriterPool{
	MaxCap: -1,
}
