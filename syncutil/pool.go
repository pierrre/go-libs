// Package syncutil provides utilities for the sync package.
package syncutil

import (
	"sync"
)

// PoolFor is a typed wrapper around [sync.Pool].
type PoolFor[T any] struct {
	p   sync.Pool
	New func() T
}

// Get is a wrapper around [sync.Pool.Get].
func (p *PoolFor[T]) Get() T {
	vi := p.p.Get()
	if vi != nil {
		return vi.(T) //nolint:forcetypeassert // The pool is typed.
	}
	if p.New != nil {
		return p.New()
	}
	var zero T
	return zero
}

// Put is a wrapper around [sync.Pool.Put].
func (p *PoolFor[T]) Put(v T) {
	p.p.Put(v)
}

// ValuePool is a [PoolFor] that works with normal value (non pointer) types.
type ValuePool[T any] struct {
	p       PoolFor[*valuePoolEntry[T]]
	entries PoolFor[*valuePoolEntry[T]]
	New     func() T
}

type valuePoolEntry[T any] struct {
	v T
}

// Get returns a value from the pool.
func (p *ValuePool[T]) Get() (v T) {
	e := p.p.Get()
	if e == nil {
		if p.New != nil {
			v = p.New()
		}
	} else {
		v, e.v = e.v, v
		p.entries.Put(e)
	}
	return v
}

// Put puts a value into the pool.
func (p *ValuePool[T]) Put(v T) {
	e := p.entries.Get()
	if e == nil {
		e = &valuePoolEntry[T]{}
	}
	e.v = v
	p.p.Put(e)
}
