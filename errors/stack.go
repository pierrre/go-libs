package errors

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"sync"
)

// WithStack adds a stack to an error.
func WithStack(err error) error {
	return withStack(err, 2)
}

func withStack(err error, skip int) error {
	if err == nil {
		return nil
	}
	return &stack{
		err:     err,
		callers: callers(skip + 1),
	}
}

type stack struct {
	err     error
	callers []uintptr
}

func (err *stack) StackFrames() *runtime.Frames {
	return runtime.CallersFrames(err.callers)
}

func (err *stack) WriteErrorMessage(w io.Writer, verbose bool) bool {
	if !verbose {
		return false
	}
	_, _ = io.WriteString(w, "stack")
	fs := err.StackFrames()
	for more := true; more; {
		var f runtime.Frame
		f, more = fs.Next()
		_, file := filepath.Split(f.File)
		_, _ = fmt.Fprintf(w, "\n\t%s %s:%d", f.Function, file, f.Line)
	}
	return true
}

func (err *stack) Error() string                 { return Error(err) }
func (err *stack) Format(s fmt.State, verb rune) { Format(err, s, verb) }
func (err *stack) Unwrap() error                 { return err.err }

// StackFrames returns the list of runtime.Frames associated to an error.
func StackFrames(err error) []*runtime.Frames {
	var fss []*runtime.Frames
	for ; err != nil; err = Unwrap(err) {
		err, ok := err.(*stack) //nolint:errorlint // We want to compare the current error.
		if ok {
			fs := err.StackFrames()
			fss = append(fss, fs)
		}
	}
	return fss
}

func ensureStack(err error, skip int) error {
	if !hasStack(err) {
		err = withStack(err, skip+1)
	}
	return err
}

func hasStack(err error) bool {
	var werr *stack
	return As(err, &werr)
}

const callersMaxLength = 1 << 16

var callersPool = sync.Pool{
	New: func() interface{} {
		return make([]uintptr, callersMaxLength)
	},
}

func callers(skip int) []uintptr {
	pcItf := callersPool.Get()
	defer callersPool.Put(pcItf)
	pc := pcItf.([]uintptr) //nolint:forcetypeassert // The pool always contains []uintptr.
	n := runtime.Callers(skip+1, pc)
	pcRes := make([]uintptr, n)
	copy(pcRes, pc)
	return pcRes
}
