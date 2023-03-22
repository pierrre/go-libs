package syncutil

import (
	"sync"
)

// Pool is a generic sync.Pool.
type Pool[T any] struct {
	pool sync.Pool

	// See sync.Pool.New.
	New func() *T
}

// Get is a wrapper for sync.Pool.Get.
func (p *Pool[T]) Get() (v *T) {
	vi := p.pool.Get()
	if vi != nil {
		return vi.(*T) //nolint:forcetypeassert // This is always a T.
	}
	if p.New != nil {
		return p.New()
	}
	return nil
}

// Put is a wrapper for sync.Pool.Put.
func (p *Pool[T]) Put(v *T) {
	p.pool.Put(v)
}
