package chaniter

import (
	"slices"
	"testing"

	"github.com/pierrre/assert"
)

func TestChan(t *testing.T) {
	ss := []string{"a", "b", "c"}
	seq := slices.All(ss)
	ch := Chan(func(yield func(e string) bool) {
		for _, e := range seq {
			if !yield(e) {
				break
			}
		}
	})
	res := make([]string, 0, len(ss))
	for e := range ch {
		res = append(res, e)
	}
	assert.DeepEqual(t, res, ss)
}

func TestSeq(t *testing.T) {
	ss := []string{"a", "b", "c"}
	ch := make(chan string, len(ss))
	for _, e := range ss {
		ch <- e
	}
	close(ch)
	seq := Seq(ch)
	res := make([]string, 0, len(ss))
	for e := range seq {
		res = append(res, e)
	}
	assert.DeepEqual(t, res, ss)
}

func TestSeqStop(t *testing.T) {
	ss := []string{"a", "b", "c"}
	ch := make(chan string, len(ss))
	for _, e := range ss {
		ch <- e
	}
	close(ch)
	seq := Seq(ch)
	for range seq {
		break
	}
}
