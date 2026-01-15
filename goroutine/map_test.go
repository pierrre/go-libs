package goroutine

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/pierrre/assert"
)

func ExampleMap() {
	ctx := context.Background()
	m := map[int]int{
		1: 1,
		2: 2,
		3: 3,
		4: 4,
		5: 5,
	}
	out := Map(ctx, m, 2, func(ctx context.Context, k int, v int) int {
		return v * 2
	})
	fmt.Println(out)
	// Output:
	// map[1:2 2:4 3:6 4:8 5:10]
}

func ExampleMapError() {
	ctx := context.Background()
	m := map[int]int{
		1: 1,
		2: 2,
		3: 3,
		4: 4,
		5: 5,
	}
	out, err := MapError(ctx, m, 2, func(ctx context.Context, k int, v int) (int, error) {
		if v == 3 {
			return 0, errors.New("error")
		}
		return v * 2, nil
	})
	fmt.Println(out)
	fmt.Println(err)
	// Output:
	// map[1:2 2:4 3:0 4:8 5:10]
	// error
}

func TestMap(t *testing.T) {
	runIterTest(t, func(t *testing.T) { //nolint:thelper // This is not a helper.
		ctx := t.Context()
		in := map[int]int{
			1: 1,
			2: 2,
			3: 3,
			4: 4,
			5: 5,
		}
		workers := 2
		f := func(ctx context.Context, k int, v int) int {
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
	})
}

func BenchmarkMap(b *testing.B) {
	ctx := b.Context()
	in := make(map[int]int)
	for i := range 100 {
		in[i] = i
	}
	f := func(ctx context.Context, k int, v int) int {
		return v * 2
	}
	for _, workers := range []int{1, 2, 5, 10} {
		b.Run(strconv.Itoa(workers), func(b *testing.B) {
			for b.Loop() {
				Map(ctx, in, workers, f)
			}
		})
	}
}

func TestMapError(t *testing.T) {
	runIterTest(t, func(t *testing.T) { //nolint:thelper // This is not a helper.
		ctx := t.Context()
		in := map[int]int{
			1: 1,
			2: 2,
			3: 3,
			4: 4,
			5: 5,
		}
		workers := 2
		f := func(ctx context.Context, k int, v int) (int, error) {
			if v == 3 {
				return 0, errors.New("error")
			}
			return v * 2, nil
		}
		out, err := MapError(ctx, in, workers, f)
		expected := map[int]int{
			1: 2,
			2: 4,
			3: 0,
			4: 8,
			5: 10,
		}
		assert.MapEqual(t, out, expected)
		assert.Error(t, err)
	})
}

func BenchmarkMapError(b *testing.B) {
	ctx := b.Context()
	in := make(map[int]int)
	for i := range 100 {
		in[i] = i
	}
	f := func(ctx context.Context, k int, v int) (int, error) {
		if k%10 == 0 {
			return 0, errors.New("error")
		}
		return v * 2, nil
	}
	for _, workers := range []int{1, 2, 5, 10} {
		b.Run(strconv.Itoa(workers), func(b *testing.B) {
			for b.Loop() {
				_, _ = MapError(ctx, in, workers, f)
			}
		})
	}
}
