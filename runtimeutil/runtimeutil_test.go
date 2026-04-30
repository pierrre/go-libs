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

var testSink any

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

func TestAppendCallersFrames(t *testing.T) {
	depth := 100
	callWithDepth(depth, func() {
		pc := GetCallers(0)
		var dst []byte
		dst = AppendCallersFrames(dst, pc)
		assert.SliceNotEmpty(t, dst)
	})
}

func TestAppendCallersFramesAllocs(t *testing.T) {
	depth := 100
	callWithDepth(depth, func() {
		pc := GetCallers(0)
		var dst []byte
		assert.AllocsPerRun(t, 100, func() {
			dst = AppendCallersFrames(dst[:0], pc)
		}, 1)
		testSink = dst
	})
}

func BenchmarkAppendCallersFrames(b *testing.B) {
	depth := 100
	callWithDepth(depth, func() {
		pc := GetCallers(0)
		var dst []byte
		for b.Loop() {
			dst = AppendCallersFrames(dst[:0], pc)
		}
	})
}

func TestWriteCallersFrames(t *testing.T) {
	depth := 100
	callWithDepth(depth, func() {
		pc := GetCallers(0)
		buf := new(bytes.Buffer)
		n, err := WriteCallersFrames(buf, pc)
		assert.NoError(t, err)
		assert.NotZero(t, n)
		assert.SliceNotEmpty(t, buf.Bytes())
	})
}

func TestWriteCallersFramesAllocs(t *testing.T) {
	depth := 100
	callWithDepth(depth, func() {
		pc := GetCallers(0)
		buf := new(bytes.Buffer)
		assert.AllocsPerRun(t, 100, func() {
			_, _ = WriteCallersFrames(buf, pc)
		}, 1)
	})
}

func BenchmarkWriteCallersFrames(b *testing.B) {
	depth := 100
	callWithDepth(depth, func() {
		pc := GetCallers(0)
		buf := new(bytes.Buffer)
		for b.Loop() {
			_, _ = WriteCallersFrames(buf, pc)
			buf.Reset()
		}
	})
}

var testFrame = runtime.Frame{
	Function: "function",
	File:     "file.go",
	Line:     123,
}

func TestAppendFrames(t *testing.T) {
	fs := slices.Values(slices.Repeat([]runtime.Frame{testFrame}, 100))
	dst := AppendFrames(nil, fs)
	assertauto.Equal(t, string(dst))
}

func TestAppendFramesAllocs(t *testing.T) {
	fs := slices.Values(slices.Repeat([]runtime.Frame{testFrame}, 100))
	var dst []byte
	assert.AllocsPerRun(t, 100, func() {
		dst = AppendFrames(dst[:0], fs)
	}, 0)
	testSink = dst
}

func BenchmarkAppendFrames(b *testing.B) {
	fs := slices.Values(slices.Repeat([]runtime.Frame{testFrame}, 100))
	var dst []byte
	for b.Loop() {
		dst = AppendFrames(dst[:0], fs)
	}
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
	}, 1)
	testSink = n
	testSink = err
}

func TestWriteFramesError(t *testing.T) {
	w := &testErrorWriter{}
	fs := slices.Values(slices.Repeat([]runtime.Frame{testFrame}, 100))
	n, err := WriteFrames(w, fs)
	assert.Error(t, err)
	assert.Equal(t, n, 0)
}

func BenchmarkWriteFrames(b *testing.B) {
	fs := slices.Values(slices.Repeat([]runtime.Frame{testFrame}, 100))
	for b.Loop() {
		_, _ = WriteFrames(io.Discard, fs)
	}
}

func TestAppendFrame(t *testing.T) {
	dst := AppendFrame(nil, testFrame)
	assertauto.Equal(t, string(dst))
}

func TestAppendFrameAllocs(t *testing.T) {
	var dst []byte
	assert.AllocsPerRun(t, 100, func() {
		dst = AppendFrame(dst[:0], testFrame)
	}, 0)
	testSink = dst
}

func BenchmarkAppendFrame(b *testing.B) {
	var dst []byte
	for b.Loop() {
		dst = AppendFrame(dst[:0], testFrame)
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
	testSink = n
	testSink = err
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
