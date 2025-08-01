// Package runtimeutil provides runtime utilities.
package runtimeutil

import (
	"io"
	"iter"
	"runtime"

	"github.com/pierrre/go-libs/strconvio"
	"github.com/pierrre/go-libs/syncutil"
	"github.com/pierrre/go-libs/unsafeio"
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

// WriteFrames writes an [iter.Seq] of [runtime.Frame] to a [io.Writer].
func WriteFrames(w io.Writer, frames iter.Seq[runtime.Frame]) (total int64, err error) {
	for f := range frames {
		var n int64
		n, err = WriteFrame(w, f)
		total += n
		if err != nil {
			break
		}
	}
	return total, err
}

// WriteFrame writes a [runtime.Frame] to a [io.Writer].
func WriteFrame(w io.Writer, f runtime.Frame) (int64, error) { //nolint:gocritic // runtime.Frame is large.
	var total int64
	n, err := unsafeio.WriteString(w, f.Function)
	total += int64(n)
	if err == nil {
		n, err = unsafeio.WriteString(w, "\n\t")
		total += int64(n)
	}
	if err == nil {
		n, err = unsafeio.WriteString(w, f.File)
		total += int64(n)
	}
	if err == nil {
		n, err = unsafeio.WriteString(w, ":")
		total += int64(n)
	}
	if err == nil {
		n, err = strconvio.WriteInt(w, int64(f.Line), 10)
		total += int64(n)
	}
	if err == nil {
		n, err = unsafeio.WriteString(w, "\n")
		total += int64(n)
	}
	return total, err //nolint:wrapcheck // No need to wrap.
}
