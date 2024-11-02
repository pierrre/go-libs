package ctxstack_test

import (
	"context"
	"runtime"
	"slices"
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/ctxstack"
)

func Test(t *testing.T) {
	ctx := context.Background()
	ctx = NewContext(ctx)
	ctx = NewContext(ctx)
	_ = ctx.Value("test")
	stacks := slices.Collect(FromContext(ctx))
	assert.SliceLen(t, stacks, 2)
	for _, stack := range stacks {
		assert.SliceNotEmpty(t, stack)
		f, _ := runtime.CallersFrames(stack).Next()
		assert.Equal(t, f.Function, "github.com/pierrre/go-libs/ctxstack_test.Test")
	}
}

func TestFromContextIterInterrupt(t *testing.T) {
	ctx := context.Background()
	ctx = NewContext(ctx)
	ctx = NewContext(ctx)
	count := 0
	for range FromContext(ctx) {
		count++
		break
	}
	assert.Equal(t, count, 1)
}

var benchRes any

func BenchmarkNewContext(b *testing.B) {
	ctx := context.Background()
	var res context.Context
	for range b.N {
		res = NewContext(ctx)
	}
	benchRes = res
}

func BenchmarkFromContext(b *testing.B) {
	ctx := context.Background()
	ctx = NewContext(ctx)
	var res []uintptr
	for range b.N {
		for stack := range FromContext(ctx) {
			res = stack
		}
	}
	benchRes = res
}
