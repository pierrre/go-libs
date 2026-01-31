package goroutine

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/pierrre/assert"
	"go.uber.org/goleak"
)

func ExampleSlice() {
	ctx := context.Background()
	s := []int{1, 2, 3, 4, 5}
	out := Slice(ctx, s, 2, func(ctx context.Context, i int, v int) int {
		return v * 2
	})
	fmt.Println(out)
	// Output:
	// [2 4 6 8 10]
}

func ExampleSliceError() {
	ctx := context.Background()
	s := []int{1, 2, 3, 4, 5}
	out, err := SliceError(ctx, s, 2, func(ctx context.Context, i int, v int) (int, error) {
		if v == 3 {
			return 0, errors.New("error")
		}
		return v * 2, nil
	})
	fmt.Println(out)
	fmt.Println(err)
	// Output:
	// [2 4 0 8 10]
	// error
}

func TestSlice(t *testing.T) {
	defer goleak.VerifyNone(t)
	runIterTest(t, func(t *testing.T) { //nolint:thelper // This is not a helper.
		ctx := t.Context()
		workers := 2
		f := func(ctx context.Context, i int, v int) int {
			return v * 2
		}
		out := Slice(ctx, testIterInputInts, workers, f)
		expected := []int{2, 4, 6, 8, 10, 12, 14, 16, 18, 20}
		assert.SliceEqual(t, out, expected)
	})
}

func BenchmarkSlice(b *testing.B) {
	ctx := b.Context()
	in := make([]int, 100)
	for i := range in {
		in[i] = i
	}
	f := func(ctx context.Context, i int, v int) int {
		return v * 2
	}
	for _, workers := range []int{1, 2, 5, 10} {
		b.Run(strconv.Itoa(workers), func(b *testing.B) {
			for b.Loop() {
				Slice(ctx, in, workers, f)
			}
		})
	}
}

func TestSliceError(t *testing.T) {
	defer goleak.VerifyNone(t)
	runIterTest(t, func(t *testing.T) { //nolint:thelper // This is not a helper.
		ctx := t.Context()
		workers := 2
		f := func(ctx context.Context, i int, v int) (int, error) {
			if v == 3 {
				return 0, errors.New("error")
			}
			return v * 2, nil
		}
		out, err := SliceError(ctx, testIterInputInts, workers, f)
		expected := []int{2, 4, 0, 8, 10, 12, 14, 16, 18, 20}
		assert.SliceEqual(t, out, expected)
		assert.Error(t, err)
	})
}

func BenchmarkSliceError(b *testing.B) {
	ctx := b.Context()
	in := make([]int, 100)
	for i := range in {
		in[i] = i
	}
	f := func(ctx context.Context, i int, v int) (int, error) {
		if i%10 == 0 {
			return 0, errors.New("error")
		}
		return v * 2, nil
	}
	for _, workers := range []int{1, 2, 5, 10} {
		b.Run(strconv.Itoa(workers), func(b *testing.B) {
			for b.Loop() {
				_, _ = SliceError(ctx, in, workers, f)
			}
		})
	}
}
