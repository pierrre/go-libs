// Package runtimeutil provides runtime utilities.
package runtimeutil

import (
	"iter"
	"runtime"

	"github.com/pierrre/go-libs/syncutil"
)

// GetCallers returns the callers.
// See [runtime.Callers].
func GetCallers(skip int) []uintptr {
	skip += 2 // Skip [GetCallers] and [runtime.Callers].
	pc := callersPool.Get()
	for {
		n := runtime.Callers(skip, pc)
		if n < len(pc) {
			res := make([]uintptr, n)
			copy(res, pc[:n])
			callersPool.Put(pc)
			return res
		}
		pc = make([]uintptr, 2*len(pc))
	}
}

var callersPool = syncutil.ValuePool[[]uintptr]{
	New: func() []uintptr {
		return make([]uintptr, 1024)
	},
}

// GetCallersFrames returns and [iter.Seq] of [runtime.Frame] for the given callers.
func GetCallersFrames(callers []uintptr) iter.Seq[runtime.Frame] {
	return func(yield func(runtime.Frame) bool) {
		frames := runtime.CallersFrames(callers)
		for {
			frame, more := frames.Next()
			if !yield(frame) || !more {
				return
			}
		}
	}
}
