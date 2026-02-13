// Package mapsutil provides utilities for maps.
package mapsutil

import (
	"cmp"
	"iter"
	"maps"
	"slices"
)

// Sorted returns an [iter.Seq2] of the key-value pairs of the map, sorted by keys.
// The keys must be ordered.
// If a key is deleted from the map while the sequence is being iterated, it will be skipped.
// If a key is added to the map while the sequence is being iterated, it will not be included in the sequence.
func Sorted[M ~map[K]V, K cmp.Ordered, V any](m M) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		keys := make([]K, 0, len(m))
		keys = slices.AppendSeq(keys, maps.Keys(m))
		slices.Sort(keys)
		for _, k := range keys {
			v, ok := m[k]
			if !ok {
				continue
			}
			if !yield(k, v) {
				return
			}
		}
	}
}
