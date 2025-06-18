// Package funcutil provides utility functions for working with functions.
package funcutil

import (
	"bytes"
	"fmt"
	"runtime/debug"
)

// Call calls the function f, then calls after with the result.
// The goexit flag indicates whether [runtime.Goexit] was called.
// The panicErr error indicates whether a panic occurred.
func Call(f func(), after func(goexit bool, panicErr error)) {
	normalReturn := false
	recovered := false
	var goexit bool
	var panicErr error
	defer func() {
		if !normalReturn && !recovered {
			goexit = true
		}
		after(goexit, panicErr)
	}()
	func() {
		defer func() {
			if !normalReturn {
				r := recover()
				if r != nil {
					panicErr = newPanicError(r)
				}
			}
		}()
		f()
	}()
	if !normalReturn {
		recovered = true
	}
}

type panicError struct {
	r     any
	stack []byte
}

func newPanicError(r any) error {
	stack := debug.Stack()
	// The first line of the stack trace is of the form "goroutine N [status]:"
	// but by the time the panic reaches Do the goroutine may no longer exist
	// and its status will have changed. Trim out the misleading line.
	if line := bytes.IndexByte(stack, '\n'); line >= 0 {
		stack = stack[line+1:]
	}
	return &panicError{r: r, stack: stack}
}

func (p *panicError) Error() string {
	return fmt.Sprintf("%v\n\n%s", p.r, p.stack)
}

func (p *panicError) Unwrap() error {
	err, _ := p.r.(error)
	return err
}
