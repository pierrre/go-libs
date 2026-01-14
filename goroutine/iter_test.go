package goroutine

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/pierrre/assert"
	"github.com/pierrre/go-libs/iterutil"
)

func ExampleIter() {
	ctx := context.Background()
	in := slices.Values([]int{1, 2, 3, 4, 5})
	out := Iter(ctx, in, 2, func(ctx context.Context, v int) int {
		return v * 2
	})
	for v := range out {
		fmt.Println(v)
	}
	// Unordered output:
	// 2
	// 4
	// 6
	// 8
	// 10
}

func ExampleIterOrdered() {
	ctx := context.Background()
	in := slices.Values([]int{1, 2, 3, 4, 5})
	out := IterOrdered(ctx, in, 2, func(ctx context.Context, v int) int {
		return v * 2
	})
	for v := range out {
		fmt.Println(v)
	}
	// Output:
	// 2
	// 4
	// 6
	// 8
	// 10
}

func ExampleIter2() {
	ctx := context.Background()
	in := slices.All([]int{1, 2, 3, 4, 5})
	out := Iter2(ctx, in, 2, func(ctx context.Context, kv iterutil.KeyVal[int, int]) int {
		return kv.Val * 2
	})
	for i, v := range out {
		fmt.Println(i, v)
	}
	// Unordered output:
	// 0 2
	// 1 4
	// 2 6
	// 3 8
	// 4 10
}

func ExampleIter2Ordered() {
	ctx := context.Background()
	in := slices.All([]int{1, 2, 3, 4, 5})
	out := Iter2Ordered(ctx, in, 2, func(ctx context.Context, kv iterutil.KeyVal[int, int]) int {
		return kv.Val * 2
	})
	for i, v := range out {
		fmt.Println(i, v)
	}
	// Output:
	// 0 2
	// 1 4
	// 2 6
	// 3 8
	// 4 10
}

func ExampleWithError() {
	ctx := context.Background()
	in := slices.Values([]int{1, 2, 3, 4, 5})
	out := Iter(ctx, in, 2, WithError(func(ctx context.Context, v int) (int, error) {
		if v == 3 {
			return 0, errors.New("error")
		}
		return v * 2, nil
	}))
	for v := range out {
		if v.Err != nil {
			fmt.Println(v.Err)
		} else {
			fmt.Println(v.Val)
		}
	}
	// Unordered output:
	// 2
	// 4
	// error
	// 8
	// 10
}

const (
	testIterCount = 100
)

var testIterInputInts = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

func runIterTest(t *testing.T, f func(t *testing.T)) {
	t.Helper()
	for range testIterCount {
		f(t)
	}
}

func TestIter(t *testing.T) {
	ctx := t.Context()
	in := slices.Values(testIterInputInts)
	workers := 2
	f := func(ctx context.Context, v int) int {
		return v * 2
	}
	out := Iter(ctx, in, workers, f)
	runIterTest(t, func(t *testing.T) { //nolint:thelper // This is not a helper.
		res := slices.Collect(out)
		slices.Sort(res)
		expected := []int{2, 4, 6, 8, 10, 12, 14, 16, 18, 20}
		assert.SliceEqual(t, res, expected)
	})
}

func TestIterStop(t *testing.T) {
	runIterTest(t, func(t *testing.T) { //nolint:thelper // This is not a helper.
		ctx := t.Context()
		in := slices.Values(testIterInputInts)
		workers := 2
		workerCallcount := int64(0)
		f := func(ctx context.Context, v int) int {
			atomic.AddInt64(&workerCallcount, 1)
			return v * 2
		}
		out := Iter(ctx, in, workers, f)
		iterCount := 0
		for range out {
			if iterCount >= 1 {
				break
			}
			iterCount++
		}
		assert.LessOrEqual(t, workerCallcount, int64(len(testIterInputInts)))
		assert.Equal(t, iterCount, 1)
	})
}

func TestIterContextCancel(t *testing.T) {
	runIterTest(t, func(t *testing.T) { //nolint:thelper // This is not a helper.
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		in := slices.Values(testIterInputInts)
		workers := 2
		workerCallcount := int64(0)
		f := func(ctx context.Context, v int) int {
			atomic.AddInt64(&workerCallcount, 1)
			return v * 2
		}
		out := Iter(ctx, in, workers, f)
		iterCount := int64(0)
		for range out {
			cancel()
			iterCount++
		}
		assert.LessOrEqual(t, workerCallcount, int64(len(testIterInputInts)))
		assert.LessOrEqual(t, iterCount, int64(len(testIterInputInts)))
		assert.Equal(t, iterCount, workerCallcount)
	})
}

func TestIterPanicInput(t *testing.T) {
	runIterTest(t, func(t *testing.T) { //nolint:thelper // This is not a helper.
		ctx := t.Context()
		in := func(func(int) bool) {
			panic("panic")
		}
		workers := 2
		f := func(ctx context.Context, v int) int {
			t.Fatal("should not be called")
			return 0
		}
		out := Iter(ctx, in, workers, f)
		iterCount := 0
		assert.Panics(t, func() {
			for range out {
				iterCount++
			}
		})
		assert.Equal(t, iterCount, 0)
	})
}

func TestIterPanicFunction(t *testing.T) {
	runIterTest(t, func(t *testing.T) { //nolint:thelper // This is not a helper.
		ctx := t.Context()
		in := slices.Values(testIterInputInts)
		workers := 2
		f := func(ctx context.Context, v int) int {
			panic("panic")
		}
		out := Iter(ctx, in, workers, f)
		iterCount := 0
		assert.Panics(t, func() {
			for range out {
				iterCount++
			}
		})
		assert.Equal(t, iterCount, 0)
	})
}

func TestIterPanicOutput(t *testing.T) {
	runIterTest(t, func(t *testing.T) { //nolint:thelper // This is not a helper.
		ctx := t.Context()
		in := slices.Values(testIterInputInts)
		workers := 2
		workerCallcount := int64(0)
		f := func(ctx context.Context, v int) int {
			atomic.AddInt64(&workerCallcount, 1)
			return v * 2
		}
		out := Iter(ctx, in, workers, f)
		assert.Panics(t, func() {
			for range out {
				panic("panic")
			}
		})
		assert.LessOrEqual(t, workerCallcount, int64(len(testIterInputInts)))
	})
}

func BenchmarkIter(b *testing.B) {
	ctx := b.Context()
	in := func(yield func(int) bool) {
		for i := range 100 {
			if !yield(i) {
				return
			}
		}
	}
	f := func(ctx context.Context, v int) int {
		return v * 2
	}
	for _, workers := range []int{1, 2, 5, 10} {
		b.Run(strconv.Itoa(workers), func(b *testing.B) {
			for b.Loop() {
				out := Iter(ctx, in, workers, f)
				out(func(int) bool {
					return true
				})
			}
		})
	}
}

func TestIterOrdered(t *testing.T) {
	ctx := t.Context()
	in := slices.Values(testIterInputInts)
	workers := 2
	f := func(ctx context.Context, v int) int {
		return v * 2
	}
	out := IterOrdered(ctx, in, workers, f)
	runIterTest(t, func(t *testing.T) { //nolint:thelper // This is not a helper.
		res := slices.Collect(out)
		expected := []int{2, 4, 6, 8, 10, 12, 14, 16, 18, 20}
		assert.SliceEqual(t, res, expected)
	})
}

func TestIterOrderedStop(t *testing.T) {
	runIterTest(t, func(t *testing.T) { //nolint:thelper // This is not a helper.
		ctx := t.Context()
		in := slices.Values(testIterInputInts)
		workers := 2
		workerCallcount := int64(0)
		f := func(ctx context.Context, v int) int {
			atomic.AddInt64(&workerCallcount, 1)
			return v * 2
		}
		out := IterOrdered(ctx, in, workers, f)
		iterCount := 0
		for range out {
			if iterCount >= 1 {
				break
			}
			iterCount++
		}
		assert.LessOrEqual(t, workerCallcount, int64(len(testIterInputInts)))
		assert.Equal(t, iterCount, 1)
	})
}

func TestIterOrderedContextCancel(t *testing.T) {
	runIterTest(t, func(t *testing.T) { //nolint:thelper // This is not a helper.
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		in := slices.Values(testIterInputInts)
		workers := 2
		workerCallcount := int64(0)
		f := func(ctx context.Context, v int) int {
			atomic.AddInt64(&workerCallcount, 1)
			return v * 2
		}
		out := IterOrdered(ctx, in, workers, f)
		iterCount := int64(0)
		for range out {
			cancel()
			iterCount++
		}
		assert.LessOrEqual(t, workerCallcount, int64(len(testIterInputInts)))
		assert.LessOrEqual(t, iterCount, int64(len(testIterInputInts)))
		assert.Equal(t, iterCount, workerCallcount)
	})
}

func TestIterOrderedPanicinput(t *testing.T) {
	runIterTest(t, func(t *testing.T) { //nolint:thelper // This is not a helper.
		ctx := t.Context()
		in := func(func(int) bool) {
			panic("panic")
		}
		workers := 2
		f := func(ctx context.Context, v int) int {
			t.Fatal("should not be called")
			return 0
		}
		out := IterOrdered(ctx, in, workers, f)
		iterCount := 0
		assert.Panics(t, func() {
			for range out {
				iterCount++
			}
		})
		assert.Equal(t, iterCount, 0)
	})
}

func TestIterOrderedPanicFunction(t *testing.T) {
	runIterTest(t, func(t *testing.T) { //nolint:thelper // This is not a helper.
		ctx := t.Context()
		in := slices.Values(testIterInputInts)
		workers := 2
		f := func(ctx context.Context, v int) int {
			panic("panic")
		}
		out := IterOrdered(ctx, in, workers, f)
		iterCount := 0
		assert.Panics(t, func() {
			for range out {
				iterCount++
			}
		})
		assert.Equal(t, iterCount, 0)
	})
}

func TestIterOrderedPanicOutput(t *testing.T) {
	runIterTest(t, func(t *testing.T) { //nolint:thelper // This is not a helper.
		ctx := t.Context()
		in := slices.Values(testIterInputInts)
		workers := 2
		workerCallcount := int64(0)
		f := func(ctx context.Context, v int) int {
			atomic.AddInt64(&workerCallcount, 1)
			return v * 2
		}
		out := IterOrdered(ctx, in, workers, f)
		assert.Panics(t, func() {
			for range out {
				panic("panic")
			}
		})
		assert.LessOrEqual(t, workerCallcount, int64(len(testIterInputInts)))
	})
}

func BenchmarkIterOrdered(b *testing.B) {
	ctx := b.Context()
	in := func(yield func(int) bool) {
		for i := range 100 {
			if !yield(i) {
				return
			}
		}
	}
	f := func(ctx context.Context, v int) int {
		return v * 2
	}
	for _, workers := range []int{1, 2, 5, 10} {
		b.Run(strconv.Itoa(workers), func(b *testing.B) {
			for b.Loop() {
				out := IterOrdered(ctx, in, workers, f)
				out(func(int) bool {
					return true
				})
			}
		})
	}
}

func TestWithError(t *testing.T) {
	runIterTest(t, func(t *testing.T) { //nolint:thelper // This is not a helper.
		ctx := t.Context()
		in := slices.Values(testIterInputInts)
		workers := 2
		f := WithError(func(ctx context.Context, v int) (int, error) {
			if v == 3 {
				return 0, errors.New("error")
			}
			return v * 2, nil
		})
		out := Iter(ctx, in, workers, f)
		errCount := 0
		for v := range out {
			if v.Val == 0 {
				errCount++
				assert.Error(t, v.Err)
			} else {
				assert.NoError(t, v.Err)
			}
		}
		assert.Equal(t, errCount, 1)
	})
}
