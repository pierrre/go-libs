package syncutil

import (
	"sync"
)

// Pool is a generic sync.Pool.
type Pool[T any] struct {
	once sync.Once
	pool sync.Pool

	// See sync.Pool.New.
	New func() T
}

func (p *Pool[T]) ensureInit() {
	p.once.Do(p.init)
}

func (p *Pool[T]) init() {
	if p.New != nil {
		p.pool.New = func() any {
			return p.New()
		}
	}
}

// Get is a wrapper for sync.Pool.Get.
//
// Instead of returning nil, it returns ok=false.
func (p *Pool[T]) Get() (v T, ok bool) {
	p.ensureInit()
	vi := p.pool.Get()
	if vi == nil {
		return v, false
	}
	v = vi.(T) //nolint:forcetypeassert // This is always a T.
	return v, true
}

// Put is a wrapper for sync.Pool.Put.
func (p *Pool[T]) Put(v T) {
	p.pool.Put(v)
}
