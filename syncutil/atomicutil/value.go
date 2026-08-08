package atomicutil

import (
	"sync/atomic"
)

type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

// Value is a typed wrapper around [atomic.Value].
//
// For primitive types and pointers, prefer the dedicated types of the [sync/atomic] package ([atomic.Int32], [atomic.Int64], [atomic.Bool], [atomic.Uint32], [atomic.Uint64], [atomic.Pointer]): they avoid interface boxing and are more efficient.
// Value is useful for arbitrary types that have no dedicated atomic equivalent, especially when [Value.CompareAndSwap] with value equality is needed.
type Value[T any] struct {
	_ noCopy
	v atomic.Value
}

// CompareAndSwap is a wrapper around [atomic.Value.CompareAndSwap].
// It panics if T is not comparable (e.g. slice, map, func).
// Unlike [atomic.Value.CompareAndSwap], it returns false if the Value is empty: the first value must be set with [Value.Store] or [Value.Swap].
func (v *Value[T]) CompareAndSwap(oldValue, newValue T) (swapped bool) {
	return v.v.CompareAndSwap(oldValue, newValue)
}

// Load is a wrapper around [atomic.Value.Load].
func (v *Value[T]) Load() T {
	vi := v.v.Load()
	value, _ := vi.(T)
	return value
}

// Store is a wrapper around [atomic.Value.Store].
func (v *Value[T]) Store(val T) {
	v.v.Store(val)
}

// Swap is a wrapper around [atomic.Value.Swap].
func (v *Value[T]) Swap(newValue T) (old T) {
	vi := v.v.Swap(newValue)
	old, _ = vi.(T)
	return old
}
