package syncutil

import (
	"iter"
)

// Set is a concurrency-safe set of keys, backed by [Map].
// The zero value is ready to use.
type Set[K comparable] struct {
	m Map[K, struct{}]
}

// Add adds key to the set.
// It reports whether key was newly added.
func (s *Set[K]) Add(key K) (added bool) {
	_, loaded := s.m.LoadOrStore(key, struct{}{})
	return !loaded
}

// All returns an iterator over the elements of the set.
// See [Map.Range] for more details.
func (s *Set[K]) All() iter.Seq[K] {
	return s.Range
}

// Clear removes all elements from the set.
func (s *Set[K]) Clear() {
	s.m.Clear()
}

// Contains reports whether key is in the set.
func (s *Set[K]) Contains(key K) bool {
	_, ok := s.m.Load(key)
	return ok
}

// Delete removes key from the set.
// It reports whether key was present.
func (s *Set[K]) Delete(key K) (deleted bool) {
	_, deleted = s.m.LoadAndDelete(key)
	return deleted
}

// Range calls f for each element in the set.
// It stops early if f returns false.
// See [Map.Range] for more details.
func (s *Set[K]) Range(f func(key K) bool) {
	s.m.Range(func(key K, _ struct{}) bool {
		return f(key)
	})
}
