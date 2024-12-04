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
	"github.com/pierrre/go-libs/panichandle"
)

var benchRes any

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

func ExampleSlice() {
	ctx := context.Background()
	out := Slice(ctx, []int{1, 2, 3, 4, 5}, 2, func(ctx context.Context, v int) int {
		return v * 2
	})
	fmt.Println(out)
	// Output:
	// [2 4 6 8 10]
}

func ExampleMap() {
	ctx := context.Background()
	out := Map(ctx, map[int]int{
		1: 1,
		2: 2,
		3: 3,
		4: 4,
		5: 5,
	}, 2, func(ctx context.Context, v int) int {
		return v * 2
	})
	fmt.Println(out)
	// Output:
	// map[1:2 2:4 3:6 4:8 5:10]
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
		assert.LessOrEqual(t, workerCallcount, int64(len(testIterInputInts)))
		assert.Equal(t, iterCount, 1)
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
		assert.LessOrEqual(t, workerCallcount, int64(len(testIterInputInts)))
	})
}

func TestIterPanicFunction(t *testing.T) {
	runIterTest(t, func(t *testing.T) { //nolint:thelper // This is not a helper.
		ctx := context.Background()
		panicCount := int64(0)
		ctx = panichandle.SetHandlerToContext(ctx, func(ctx context.Context, r any) {
			atomic.AddInt64(&panicCount, 1)
		})
		in := slices.Values(testIterInputInts)
		workers := 2
		f := func(ctx context.Context, v int) int {
			panic("panic")
		}
		out := Iter(ctx, in, workers, f)
		iterCount := 0
		for range out {
			iterCount++
		}
		assert.Equal(t, panicCount, int64(len(testIterInputInts)))
		assert.Equal(t, iterCount, 0)
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

func TestIterOrdered(t *testing.T) {
	ctx := context.Background()
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

func TestIterOrderedStopOutputIterator(t *testing.T) {
	runIterTest(t, func(t *testing.T) { //nolint:thelper // This is not a helper.
		ctx := context.Background()
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
		ctx, cancel := context.WithCancel(context.Background())
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

func TestIterOrderedPanicIterator(t *testing.T) {
	runIterTest(t, func(t *testing.T) { //nolint:thelper // This is not a helper.
		ctx := context.Background()
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

func TestIterOrderedPanicFunction(t *testing.T) {
	runIterTest(t, func(t *testing.T) { //nolint:thelper // This is not a helper.
		ctx := context.Background()
		panicCount := int64(0)
		ctx = panichandle.SetHandlerToContext(ctx, func(ctx context.Context, r any) {
			atomic.AddInt64(&panicCount, 1)
		})
		in := slices.Values(testIterInputInts)
		workers := 2
		f := func(ctx context.Context, v int) int {
			panic("panic")
		}
		out := IterOrdered(ctx, in, workers, f)
		iterCount := 0
		for range out {
			iterCount++
		}
		assert.Equal(t, panicCount, int64(len(testIterInputInts)))
		assert.Equal(t, iterCount, 0)
	})
}

func BenchmarkIterOrdered(b *testing.B) {
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
				out := IterOrdered(ctx, in, workers, f)
				for range out {
				}
			}
		})
	}
}

func TestSlice(t *testing.T) {
	ctx := context.Background()
	workers := 2
	f := func(ctx context.Context, v int) int {
		return v * 2
	}
	out := Slice(ctx, testIterInputInts, workers, f)
	expected := []int{2, 4, 6, 8, 10, 12, 14, 16, 18, 20}
	assert.SliceEqual(t, out, expected)
}

func BenchmarkSlice(b *testing.B) {
	ctx := context.Background()
	in := make([]int, 100)
	for i := range in {
		in[i] = i
	}
	f := func(ctx context.Context, v int) int {
		return v * 2
	}
	b.ResetTimer()
	var res []int
	for _, workers := range []int{1, 2, 5, 10} {
		b.Run(strconv.Itoa(workers), func(b *testing.B) {
			for range b.N {
				out := Slice(ctx, in, workers, f)
				res = out
			}
		})
	}
	benchRes = res
}

func TestMap(t *testing.T) {
	ctx := context.Background()
	in := map[int]int{
		1: 1,
		2: 2,
		3: 3,
		4: 4,
		5: 5,
	}
	workers := 2
	f := func(ctx context.Context, v int) int {
		return v * 2
	}
	out := Map(ctx, in, workers, f)
	expected := map[int]int{
		1: 2,
		2: 4,
		3: 6,
		4: 8,
		5: 10,
	}
	assert.MapEqual(t, out, expected)
}

func BenchmarkMap(b *testing.B) {
	ctx := context.Background()
	in := make(map[int]int)
	for i := range 100 {
		in[i] = i
	}
	f := func(ctx context.Context, v int) int {
		return v * 2
	}
	b.ResetTimer()
	var res map[int]int
	for _, workers := range []int{1, 2, 5, 10} {
		b.Run(strconv.Itoa(workers), func(b *testing.B) {
			for range b.N {
				out := Map(ctx, in, workers, f)
				res = out
			}
		})
	}
	benchRes = res
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

func TestCancelOnError(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	f := func(ctx context.Context, v int) (int, error) {
		if v == 3 {
			return 0, errors.New("error")
		}
		return v * 2, nil
	}
	f = CancelOnError(cancel, f)
	_, err := f(ctx, 0)
	assert.NoError(t, err)
	assert.NoError(t, ctx.Err())
	_, err = f(ctx, 3)
	assert.Error(t, err)
	assert.ErrorIs(t, ctx.Err(), context.Canceled)
}
