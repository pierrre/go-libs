// Package chaniter provides utilities for iterator and chan.
package chaniter

import (
	"iter"
)

// Chan returns a channel that sends elements from the given sequence.
func Chan[E any](seq iter.Seq[E]) <-chan E {
	ch := make(chan E)
	go func() {
		defer close(ch)
		for e := range seq {
			ch <- e
		}
	}()
	return ch
}

// Seq returns a sequence that sends elements from the given channel.
func Seq[E any](ch <-chan E) iter.Seq[E] {
	return func(yield func(E) bool) {
		for e := range ch {
			if !yield(e) {
				break
			}
		}
	}
}
