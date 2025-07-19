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
					panicErr = panicutil.NewError(r)
				}
			}
		}()
		f()
	}()
	if !normalReturn {
		recovered = true
	}
}
