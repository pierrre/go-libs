// Package mapsutil provides utilities for maps.
package mapsutil

import (
	"cmp"
	"iter"
	"maps"
	"slices"
)

// SortedByKey returns an [iter.Seq2] of the key-value pairs of the map, sorted by ordered keys .
// If a key is deleted from the map while the sequence is being iterated, it will be skipped.
// If a key is added to the map while the sequence is being iterated, it will not be included in the sequence.
func SortedByKey[M ~map[K]V, K cmp.Ordered, V any](m M) iter.Seq2[K, V] {
	return sortedByKey(m, slices.Sort)
}

// SortedByKeyFunc is like [SortedByKey] but the keys are sorted using the provided comparison function.
func SortedByKeyFunc[M ~map[K]V, K comparable, V any](m M, cmpFunc func(K, K) int) iter.Seq2[K, V] {
	return sortedByKey(m, func(keys []K) {
		slices.SortFunc(keys, cmpFunc)
	})
}

func sortedByKey[M ~map[K]V, K comparable, V any](m M, sortFunc func([]K)) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		keys := make([]K, 0, len(m))
		keys = slices.AppendSeq(keys, maps.Keys(m))
		sortFunc(keys)
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
