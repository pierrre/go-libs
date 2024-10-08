package syncutil

import (
	"sync"
)

// Pool is a typed wrapper around [sync.Pool].
type Pool[T any] struct {
	p   sync.Pool
	New func() T
}

// Get is a wrapper around [sync.Pool.Get].
func (p *Pool[T]) Get() T {
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
func (p *Pool[T]) Put(v T) {
	p.p.Put(v)
}

// ValuePool is a [Pool] that works with normal value (non pointer) types.
type ValuePool[T any] struct {
	p       Pool[*valuePoolEntry[T]]
	entries Pool[*valuePoolEntry[T]]
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
