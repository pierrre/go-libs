package chansutil

import (
	"context"
	"slices"
	"testing"

	"github.com/pierrre/assert"
)

func TestIter(t *testing.T) {
	ch := make(chan int, 3)
	ch <- 1
	ch <- 2
	ch <- 3
	close(ch)
	it := Iter(ch)
	for e := range it {
		assert.Equal(t, e, 1)
		break
	}
	res := slices.Collect(it)
	expected := []int{2, 3}
	assert.SliceEqual(t, res, expected)
}

func TestCollectTo(t *testing.T) {
	ctx := context.Background()
	vs := []int{1, 2, 3}
	it := slices.Values(vs)
	ch := make(chan int, len(vs))
	err := CollectTo(ctx, it, ch)
	assert.NoError(t, err)
	close(ch)
	res := slices.Collect(Iter(ch))
	assert.SliceEqual(t, res, vs)
}

func TestCollectToContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	vs := []int{1, 2, 3}
	it := slices.Values(vs)
	ch := make(chan int)
	err := CollectTo(ctx, it, ch)
	assert.ErrorIs(t, err, context.Canceled)
}
