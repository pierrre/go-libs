// Package chansutil provides utility functions for working with channels.
package chansutil

import (
	"context"
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

// CollectTo collects elements from an [iter.Seq] and sends them to a channel.
// If the context is cancelled, it stops collecting and returns the context error.
func CollectTo[E any](ctx context.Context, it iter.Seq[E], ch chan<- E) error {
	done := ctx.Done()
	for e := range it {
		select {
		case ch <- e:
		case <-done:
		}
		select {
		case <-done:
			return ctx.Err() //nolint:wrapcheck // We want to return the original context error.
		default:
		}
	}
	return nil
}
