package runtimeutil_test

import (
	"runtime"
	"testing"

	"github.com/pierrre/assert"
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
		b.ResetTimer()
		for range b.N {
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
