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
	ctx := context.Background()
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

func TestIterStopOutputIterator(t *testing.T) {
	runIterTest(t, func(t *testing.T) { //nolint:thelper // This is not a helper.
		ctx := context.Background()
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
		assert.LessOrEqual(t, workerCallcount, 5)
	})
}

func TestIterContextCancel(t *testing.T) {
	runIterTest(t, func(t *testing.T) { //nolint:thelper // This is not a helper.
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
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
			cancel()
			iterCount++
		}
		assert.LessOrEqual(t, workerCallcount, 4)
		assert.LessOrEqual(t, iterCount, 4)
	})
}

func TestIterPanicIterator(t *testing.T) {
	runIterTest(t, func(t *testing.T) { //nolint:thelper // This is not a helper.
		ctx := context.Background()
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
		assert.LessOrEqual(t, workerCallcount, 4)
	})
}

func TestWithError(t *testing.T) {
	runIterTest(t, func(t *testing.T) { //nolint:thelper // This is not a helper.
		ctx := context.Background()
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

func BenchmarkIter(b *testing.B) {
	ctx := context.Background()
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
	b.ResetTimer()
	for _, workers := range []int{1, 2, 5, 10} {
		b.Run(strconv.Itoa(workers), func(b *testing.B) {
			for range b.N {
				out := Iter(ctx, in, workers, f)
				for range out {
				}
			}
		})
	}
}
