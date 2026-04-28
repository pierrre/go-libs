// Package runtimeutil provides runtime utilities.
package runtimeutil

import (
	"io"
	"iter"
	"runtime"
	"strconv"

	"github.com/pierrre/go-libs/bytesutil"
	"github.com/pierrre/go-libs/syncutil"
)

// GetCallers returns the callers.
// See [runtime.Callers].
func GetCallers(skip int) []uintptr {
	skip += 2 // Skip [GetCallers] and [runtime.Callers].
	pc := callersPool.Get()
	for {
		n := runtime.Callers(skip, *pc)
		if n < len(*pc) {
			res := make([]uintptr, n)
			copy(res, (*pc)[:n])
			callersPool.Put(pc)
			return res
		}
		*pc = make([]uintptr, 2*len(*pc))
	}
}

var callersPool = syncutil.Pool[*[]uintptr]{
	New: func() *[]uintptr {
		return new(make([]uintptr, 1024))
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

// WriteCallersFrames writes the frames of the given callers to a [io.Writer].
func WriteCallersFrames(w io.Writer, callers []uintptr) (total int64, err error) {
	bw := bytesWriterPool.Get()
	defer bytesWriterPool.Put(bw)
	*bw = AppendCallersFrames(*bw, callers)
	n, err := w.Write(*bw)
	return int64(n), err
}

// AppendCallersFrames appends the frames of the given callers to a []byte.
func AppendCallersFrames(dst []byte, callers []uintptr) []byte {
	return AppendFrames(dst, GetCallersFrames(callers))
}

// WriteFrames writes an [iter.Seq] of [runtime.Frame] to a [io.Writer].
func WriteFrames(w io.Writer, frames iter.Seq[runtime.Frame]) (total int64, err error) {
	bw := bytesWriterPool.Get()
	defer bytesWriterPool.Put(bw)
	frames(func(f runtime.Frame) bool {
		*bw = AppendFrame(*bw, f)
		return true
	})
	n, err := w.Write(*bw)
	return int64(n), err
}

// AppendFrames appends an [iter.Seq] of [runtime.Frame] to a []byte.
func AppendFrames(dst []byte, frames iter.Seq[runtime.Frame]) []byte {
	frames(func(f runtime.Frame) bool {
		dst = AppendFrame(dst, f)
		return true
	})
	return dst
}

// WriteFrame writes a [runtime.Frame] to a [io.Writer].
func WriteFrame(w io.Writer, f runtime.Frame) (int64, error) { //nolint:gocritic // runtime.Frame is large.
	bw := bytesWriterPool.Get()
	defer bytesWriterPool.Put(bw)
	*bw = AppendFrame(*bw, f)
	n, err := w.Write(*bw)
	return int64(n), err
}

// AppendFrame appends a [runtime.Frame] to a []byte.
func AppendFrame(dst []byte, f runtime.Frame) []byte { //nolint:gocritic // runtime.Frame is large.
	dst = append(dst, f.Function...)
	dst = append(dst, "\n\t"...)
	dst = append(dst, f.File...)
	dst = append(dst, ':')
	dst = strconv.AppendInt(dst, int64(f.Line), 10)
	dst = append(dst, '\n')
	return dst
}

var bytesWriterPool = &bytesutil.WriterPool{
	MaxCap: -1,
}
