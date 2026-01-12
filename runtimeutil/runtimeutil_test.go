package runtimeutil_test

import (
	"bytes"
	"errors"
	"io"
	"runtime"
	"slices"
	"testing"

	"github.com/pierrre/assert"
	"github.com/pierrre/assert/assertauto"
	. "github.com/pierrre/go-libs/runtimeutil"
)

func TestGetCallers(t *testing.T) {
	depth := 10000
	callWithDepth(depth, func() {
		pc := GetCallers(0)
		assert.GreaterOrEqual(t, len(pc), depth)
		fs := runtime.CallersFrames(pc)
		f, _ := fs.Next()
		assert.StringContains(t, f.Function, ".TestGetCallers.")
	})
}

func TestGetCallersAllocs(t *testing.T) {
	depth := 10000
	callWithDepth(depth, func() {
		assert.AllocsPerRun(t, 100, func() {
			_ = GetCallers(0)
		}, 1)
	})
}

func BenchmarkGetCallers(b *testing.B) {
	callWithDepth(10000, func() {
		for b.Loop() {
			pc := GetCallers(0)
			_ = pc
		}
	})
}

func callWithDepth(depth int, f func()) {
	if depth > 0 {
		callWithDepth(depth-1, f)
	} else {
		f()
	}
}

func TestGetCallersFrames(t *testing.T) {
	depth := 10
	callWithDepth(depth, func() {
		pc := GetCallers(0)
		count := 0
		for range GetCallersFrames(pc) {
			count++
		}
		assert.GreaterOrEqual(t, count, depth)
	})
}

var testFrame = runtime.Frame{
	Function: "function",
	File:     "file.go",
	Line:     123,
}

func TestWriteFrames(t *testing.T) {
	buf := new(bytes.Buffer)
	fs := slices.Values(slices.Repeat([]runtime.Frame{testFrame}, 100))
	n, err := WriteFrames(buf, fs)
	assert.NoError(t, err)
	assertauto.Equal(t, n)
	assertauto.Equal(t, buf.String())
}

func TestWriteFramesAllocs(t *testing.T) {
	fs := slices.Values(slices.Repeat([]runtime.Frame{testFrame}, 100))
	var n int64
	var err error
	assert.AllocsPerRun(t, 100, func() {
		n, err = WriteFrames(io.Discard, fs)
	}, 0)
	runtime.KeepAlive(n)
	runtime.KeepAlive(err)
}

func TestWriteFramesError(t *testing.T) {
	w := &testErrorWriter{}
	fs := slices.Values(slices.Repeat([]runtime.Frame{testFrame}, 100))
	n, err := WriteFrames(w, fs)
	assert.Error(t, err)
	assert.Equal(t, n, 0)
}

var benchRes any

func BenchmarkWriteFrames(b *testing.B) {
	fs := slices.Values(slices.Repeat([]runtime.Frame{testFrame}, 100))

	var n int64
	var err error
	for b.Loop() {
		n, err = WriteFrames(io.Discard, fs)
	}
	benchRes = n
	benchRes = err
}

func BenchmarkWriteFramesNew(b *testing.B) {
	fs := slices.Values(slices.Repeat([]runtime.Frame{testFrame}, 100))
	for b.Loop() {
		_, _ = WriteFrames(io.Discard, fs)
	}
}

func TestWriteFrame(t *testing.T) {
	buf := new(bytes.Buffer)
	n, err := WriteFrame(buf, testFrame)
	assert.NoError(t, err)
	assertauto.Equal(t, n)
	assertauto.Equal(t, buf.String())
}

func TestWriteFrameAllocs(t *testing.T) {
	var n int64
	var err error
	assert.AllocsPerRun(t, 100, func() {
		n, err = WriteFrame(io.Discard, testFrame)
	}, 0)
	runtime.KeepAlive(n)
	runtime.KeepAlive(err)
}

func TestWriteFrameError(t *testing.T) {
	w := &testErrorWriter{}
	n, err := WriteFrame(w, testFrame)
	assert.Error(t, err)
	assert.Equal(t, n, 0)
}

func BenchmarkWriteFrame(b *testing.B) {
	for b.Loop() {
		_, _ = WriteFrame(io.Discard, testFrame)
	}
}

type testErrorWriter struct{}

func (w *testErrorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("error")
}
