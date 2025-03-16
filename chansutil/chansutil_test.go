package chansutil

import (
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
