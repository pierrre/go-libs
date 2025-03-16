// Package chansutil provides utility functions for working with channels.
package chansutil

import (
	"iter"
)

// Iter returns a [iter.Seq] for a channel.
func Iter[C ~chan E, E any](ch C) iter.Seq[E] {
	return func(yield func(E) bool) {
		for e := range ch {
			if !yield(e) {
				break
			}
		}
	}
}
