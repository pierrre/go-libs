package iterutil

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/pierrre/assert"
)

func ExampleSeqToSeq2() {
	kvs := []KeyVal[int, string]{
		NewKeyVal(0, "zero"),
		NewKeyVal(1, "one"),
		NewKeyVal(2, "two"),
	}
	it := SeqToSeq2(slices.Values(kvs), KeyVal[int, string].Values)
	for k, v := range it {
		fmt.Println(k, v)
	}
	// Output:
	// 0 zero
	// 1 one
	// 2 two
}

func ExampleSeq2ToSeq() {
	ss := []string{"zero", "one", "two"}
	it := Seq2ToSeq(slices.All(ss), NewKeyVal)
	for kv := range it {
		fmt.Println(kv.Key, kv.Val)
	}
	// Output:
	// 0 zero
	// 1 one
	// 2 two
}

func TestSeqToSeq2(t *testing.T) {
	kvs := []KeyVal[int, string]{
		NewKeyVal(0, "zero"),
		NewKeyVal(1, "one"),
		NewKeyVal(2, "two"),
	}
	it := SeqToSeq2(slices.Values(kvs), KeyVal[int, string].Values)
	for k, v := range it {
		assert.Equal(t, v, kvs[k].Val)
	}
}

func TestSeqToSeq2Stop(t *testing.T) {
	kvs := []KeyVal[int, string]{
		NewKeyVal(0, "zero"),
		NewKeyVal(1, "one"),
		NewKeyVal(2, "two"),
	}
	it := SeqToSeq2(slices.Values(kvs), KeyVal[int, string].Values)
	for k, v := range it {
		assert.Equal(t, v, kvs[k].Val)
		break
	}
}

func TestSeqToSeq2Index(t *testing.T) {
	ss := []string{"zero", "one", "two"}
	it := SeqToSeq2Index(slices.Values(ss))
	for k, v := range it {
		assert.Equal(t, v, ss[k])
	}
}

func TestSeqToSeq2IndexStop(t *testing.T) {
	ss := []string{"zero", "one", "two"}
	it := SeqToSeq2Index(slices.Values(ss))
	for k, v := range it {
		assert.Equal(t, v, ss[k])
		break
	}
}

func TestSeq2ToSeq(t *testing.T) {
	ss := []string{"zero", "one", "two"}
	it := Seq2ToSeq(slices.All(ss), NewKeyVal)
	for kv := range it {
		assert.Equal(t, kv.Val, ss[kv.Key])
	}
}

func TestSeq2ToSeqStop(t *testing.T) {
	ss := []string{"zero", "one", "two"}
	it := Seq2ToSeq(slices.All(ss), NewKeyVal)
	for kv := range it {
		assert.Equal(t, kv.Val, ss[kv.Key])
		break
	}
}

func TestSeq2ToSeqKey(t *testing.T) {
	ss := []string{"zero", "one", "two"}
	it := Seq2ToSeqKey(slices.All(ss))
	expected := 0
	for k := range it {
		assert.Equal(t, k, expected)
		expected++
	}
	assert.Equal(t, expected, len(ss))
}

func TestSeq2ToSeqValue(t *testing.T) {
	ss := []string{"zero", "one", "two"}
	it := Seq2ToSeqValue(slices.All(ss))
	i := 0
	for v := range it {
		assert.Equal(t, v, ss[i])
		i++
	}
	assert.Equal(t, i, len(ss))
}

func ExampleRepeat() {
	ss := []string{"zero", "one", "two"}
	it := Repeat(slices.Values(ss), 2)
	for v := range it {
		fmt.Println(v)
	}
	// Output:
	// zero
	// one
	// two
	// zero
	// one
	// two
}

func TestRepeat(t *testing.T) {
	ss := []string{"zero", "one", "two"}
	it := Repeat(slices.Values(ss), 2)
	res := slices.Collect(it)
	expected := slices.Repeat(ss, 2)
	assert.SliceEqual(t, res, expected)
}

func TestRepeatInfinite(t *testing.T) {
	ss := []string{"zero", "one", "two"}
	it := Repeat(slices.Values(ss), -1)
	i := 0
	for range it {
		i++
		if i >= 10 {
			break
		}
	}
	assert.Equal(t, i, 10)
}

func TestRepeatEmpty(t *testing.T) {
	ss := []string{}
	it := Repeat(slices.Values(ss), 2)
	i := 0
	for range it {
		i++
	}
	assert.Equal(t, i, 0)
}

func ExampleRepeat2() {
	ss := []string{"zero", "one", "two"}
	it := Repeat2(slices.All(ss), 2)
	for k, v := range it {
		fmt.Println(k, v)
	}
	// Output:
	// 0 zero
	// 1 one
	// 2 two
	// 0 zero
	// 1 one
	// 2 two
}

func TestRepeat2(t *testing.T) {
	ss := []string{"zero", "one", "two"}
	it := Repeat2(slices.All(ss), 2)
	res := slices.Collect(Seq2ToSeq(it, NewKeyVal))
	expected := slices.Repeat(slices.Collect(Seq2ToSeq(slices.All(ss), NewKeyVal)), 2)
	assert.SliceEqual(t, res, expected)
}

func TestRepeat2Infinite(t *testing.T) {
	ss := []string{"zero", "one", "two"}
	it := Repeat2(slices.All(ss), -1)
	i := 0
	for range it {
		i++
		if i >= 10 {
			break
		}
	}
	assert.Equal(t, i, 10)
}

func TestRepeat2Empty(t *testing.T) {
	ss := []string{}
	it := Repeat2(slices.All(ss), 2)
	i := 0
	for range it {
		i++
	}
	assert.Equal(t, i, 0)
}

func ExampleWithConsumerStallTimeout() {
	errStalled := errors.New("stream stalled")
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	seq := WithConsumerStallTimeout(slices.Values([]int{0, 1, 2}), time.Minute, func() {
		cancel(errStalled)
	})
	for v := range seq {
		fmt.Println(v)
	}
	fmt.Println(context.Cause(ctx))
	// Output:
	// 0
	// 1
	// 2
	// <nil>
}

func TestWithConsumerStallTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var calls atomic.Int64
		stallTimeout := 5 * time.Millisecond
		seq := WithConsumerStallTimeout(slices.Values([]int{0, 1, 2}), stallTimeout, func() {
			calls.Add(1)
		})
		got := slices.Collect(seq)
		assert.SliceEqual(t, got, []int{0, 1, 2})
		assert.Equal(t, calls.Load(), 0)
	})
}

func TestWithConsumerStallTimeoutAlmostStall(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var calls atomic.Int64
		stallTimeout := 5 * time.Millisecond
		seq := WithConsumerStallTimeout(slices.Values([]int{0}), stallTimeout, func() {
			calls.Add(1)
		})
		count := 0
		for range seq {
			count++
			time.Sleep(stallTimeout - time.Nanosecond) // Hold the value just under the timeout.
		}
		assert.Equal(t, count, 1)
		assert.Equal(t, calls.Load(), 0)
	})
}

func TestWithConsumerStallTimeoutStall(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var calls atomic.Int64
		stallTimeout := 5 * time.Millisecond
		seq := WithConsumerStallTimeout(slices.Values([]int{0, 1, 2}), stallTimeout, func() {
			calls.Add(1)
		})
		count := 0
		for range seq {
			count++
			time.Sleep(10 * stallTimeout) // Hold the first value longer than the timeout.
		}
		assert.Equal(t, count, 1)
		assert.Equal(t, calls.Load(), 1)
	})
}

func TestWithConsumerStallTimeoutStallLastValue(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var calls atomic.Int64
		stallTimeout := 5 * time.Millisecond
		seq := WithConsumerStallTimeout(slices.Values([]int{0}), stallTimeout, func() {
			calls.Add(1)
		})
		count := 0
		for range seq {
			count++
			time.Sleep(10 * stallTimeout) // Hold the last value longer than the timeout.
		}
		assert.Equal(t, count, 1)
		assert.Equal(t, calls.Load(), 1)
	})
}

func TestWithConsumerStallTimeoutSlowSeq(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var calls atomic.Int64
		stallTimeout := 5 * time.Millisecond
		slowSeq := func(yield func(int) bool) {
			for _, v := range []int{0, 1, 2} {
				if !yield(v) {
					return
				}
				time.Sleep(10 * stallTimeout) // Producer time between values: not watched.
			}
		}
		seq := WithConsumerStallTimeout(slowSeq, stallTimeout, func() {
			calls.Add(1)
		})
		got := slices.Collect(seq)
		assert.SliceEqual(t, got, []int{0, 1, 2})
		assert.Equal(t, calls.Load(), 0)
	})
}

func TestWithConsumerStallTimeoutStop(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var calls atomic.Int64
		stallTimeout := 5 * time.Millisecond
		seq := WithConsumerStallTimeout(slices.Values([]int{0, 1, 2}), stallTimeout, func() {
			calls.Add(1)
		})
		for range seq {
			break
		}
		time.Sleep(2 * stallTimeout)
		assert.Equal(t, calls.Load(), 0)
	})
}

func TestWithConsumerStallTimeoutPanic(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var calls atomic.Int64
		stallTimeout := 5 * time.Millisecond
		seq := WithConsumerStallTimeout(slices.Values([]int{0, 1, 2}), stallTimeout, func() {
			calls.Add(1)
		})
		assert.Panics(t, func() {
			for range seq {
				panic("boom")
			}
		})
		time.Sleep(2 * stallTimeout) // A watchdog leaked by the panic would fire here.
		assert.Equal(t, calls.Load(), 0)
	})
}

func TestWithConsumerStallTimeoutNilOnStall(t *testing.T) {
	assert.Panics(t, func() {
		WithConsumerStallTimeout(slices.Values([]int{}), 5*time.Millisecond, nil)
	})
}

func TestWithConsumerStallTimeoutInvalidTimeout(t *testing.T) {
	for _, timeout := range []time.Duration{0, -5 * time.Millisecond} {
		assert.Panics(t, func() {
			WithConsumerStallTimeout(slices.Values([]int{}), timeout, func() {})
		})
	}
}

func TestWithConsumerStallTimeoutRerange(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var calls atomic.Int64
		stallTimeout := 5 * time.Millisecond
		seq := WithConsumerStallTimeout(slices.Values([]int{0, 1}), stallTimeout, func() {
			calls.Add(1)
		})
		count := 0
		for range seq {
			count++
			time.Sleep(10 * stallTimeout) // Stall the first range.
		}
		assert.Equal(t, count, 1)
		count = 0
		for range seq {
			count++
		}
		assert.Equal(t, count, 2)
		assert.Equal(t, calls.Load(), 1)
	})
}

func ExampleWithConsumerStallTimeout2() {
	it := WithConsumerStallTimeout2(slices.All([]string{"zero", "one", "two"}), time.Minute, func() {})
	for k, v := range it {
		fmt.Println(k, v)
	}
	// Output:
	// 0 zero
	// 1 one
	// 2 two
}

func TestWithConsumerStallTimeout2(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var calls atomic.Int64
		stallTimeout := 5 * time.Millisecond
		it := WithConsumerStallTimeout2(slices.All([]string{"zero", "one", "two"}), stallTimeout, func() {
			calls.Add(1)
		})
		count := 0
		for range it {
			count++
			time.Sleep(10 * stallTimeout) // Hold the first pair longer than the timeout.
		}
		assert.Equal(t, count, 1)
		assert.Equal(t, calls.Load(), 1)
	})
}
