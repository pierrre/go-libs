// Package singleflight provides a duplicate function call suppression mechanism.
//
// It is inspired by golang.org/x/sync/singleflight, and supports additionally generic types and context cancellation.
package singleflight

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/pierrre/go-libs/funcutil"
	"github.com/pierrre/go-libs/syncutil"
)

// Func represents a function used by [Group].
type Func[A any, V any] func(ctx context.Context, arg A) (V, error)

// Group deduplicates function calls for the same key.
type Group[K comparable, A any, V any] struct {
	once sync.Once
	mu   sync.Mutex
	m    map[K]*call[V]
	pool syncutil.Pool[*call[V]]
	// OnWait is called when a call for the same key is already in progress, and it's waiting for it to finish.
	OnWait func(ctx context.Context, key K)
}

func (g *Group[K, A, V]) init() {
	g.once.Do(func() {
		g.m = make(map[K]*call[V])
		g.pool.New = func() *call[V] {
			return new(call[V])
		}
	})
}

// Do calls the [Func], deduplicating calls for the same key.
// If a call for the same key is already in progress, it waits for it to finish and returns its result.
// The argument arg of the first call is passed to the [Func].
// The context cancellation is propagated to the [Func], unless it's from a waiting caller.
// Calling [runtime.Goexit] or panic from the [Func] is propagated to all callers.
func (g *Group[K, A, V]) Do(ctx context.Context, key K, arg A, f Func[A, V]) (v V, err error, shared bool) {
	g.init()
	c, exists := g.getOrCreateCall(key)
	defer func() {
		if c.releaseUsage() == 0 {
			*c = call[V]{}
			g.pool.Put(c)
		}
	}()
	if exists {
		v, err = g.waitCall(ctx, key, c)
		return v, err, true
	}
	return g.doCall(ctx, key, arg, c, f)
}

func (g *Group[K, A, V]) getOrCreateCall(key K) (c *call[V], exists bool) {
	g.mu.Lock()
	c, exists = g.m[key]
	if exists {
		if !c.doneInitialized {
			c.done = make(chan struct{})
			c.doneInitialized = true
		}
		c.shared = true
	} else {
		c = g.pool.Get()
		g.m[key] = c
	}
	c.incUsages()
	g.mu.Unlock()
	return c, exists
}

func (g *Group[K, A, V]) waitCall(ctx context.Context, key K, c *call[V]) (v V, err error) {
	if g.OnWait != nil {
		g.OnWait(ctx, key)
	}
	if c.done != nil {
		select {
		case <-c.done:
		case <-ctx.Done():
			return v, ctx.Err() //nolint:wrapcheck // Not needed.
		}
	}
	c.checkGoexit()
	c.checkPanic()
	return c.v, c.err
}

func (g *Group[K, A, V]) doCall(ctx context.Context, key K, arg A, c *call[V], f Func[A, V]) (v V, err error, shared bool) {
	funcutil.Call(
		func() {
			c.v, c.err = f(ctx, arg)
		},
		func(goexit bool, panicErr error) {
			c.goexit = goexit
			c.panicErr = panicErr
			g.mu.Lock()
			if g.m[key] == c {
				delete(g.m, key)
			}
			if !c.doneInitialized {
				c.doneInitialized = true
			}
			g.mu.Unlock()
			if c.done != nil {
				close(c.done)
			}

			c.checkPanic()
		},
	)
	return c.v, c.err, c.shared
}

// Forget forgets the call for the given key.
// Future calls to [Group.Do] for this key will call the function rather than waiting for a call to finish.
func (g *Group[K, A, V]) Forget(key K) {
	g.init()
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()
}

type call[V any] struct {
	doneInitialized bool
	done            chan struct{}
	v               V
	err             error
	shared          bool
	goexit          bool
	panicErr        error
	usages          int32
}

func (c *call[V]) incUsages() {
	atomic.AddInt32(&c.usages, 1)
}

func (c *call[V]) releaseUsage() int32 {
	return atomic.AddInt32(&c.usages, -1)
}

func (c *call[V]) checkGoexit() {
	if c.goexit {
		runtime.Goexit()
	}
}

func (c *call[V]) checkPanic() {
	if c.panicErr != nil {
		panic(c.panicErr)
	}
}
