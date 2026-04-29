// Package funcutil provides utility functions for working with functions.
package funcutil

import (
	"github.com/pierrre/go-libs/panicutil"
)

// Call calls the function f, then calls after with the result.
// The goexit flag indicates whether [runtime.Goexit] was called.
// The panicErr error indicates whether a panic occurred.
func Call(f func(), after func(goexit bool, panicErr error)) {
	normalReturn := false
	recovered := false
	var panicErr error
	defer func() {
		goexit := !normalReturn && !recovered
		after(goexit, panicErr)
	}()
	func() {
		defer func() {
			if !normalReturn {
				r := recover()
				if r != nil {
					panicErr = panicutil.NewError(r)
				}
			}
		}()
		f()
		normalReturn = true
	}()
	if !normalReturn {
		recovered = true
	}
}
