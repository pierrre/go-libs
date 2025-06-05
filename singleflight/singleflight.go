// Package singleflight provides a duplicate function call suppression mechanism.
//
// It is inspired by golang.org/x/sync/singleflight
// It supports additionally generic types and context cancellation.
package singleflight

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"

	"github.com/pierrre/go-libs/syncutil"
)

// Func represents a function used by [Group].
type Func[A any, V any] func(ctx context.Context, arg A) (V, error)

// Group deduplicates function calls for the same key.
type Group[K comparable, A any, V any] struct {
	calls calls[K, V]
	// OnWait is called when a call for the same key is already in progress, and it's waiting for it to finish.
	OnWait func(ctx context.Context, key K)
}

// Do calls the [Func] f, deduplicating calls for the same key.
// If a call for the same key is already in progress, it waits for it to finish and returns its result.
// The argument arg of the first call is passed to the [Func].
// The context cancellation is propagated to the call, unless it is already in progress.
// [runtime.Goexit] is propagated.
func (g *Group[K, A, V]) Do(ctx context.Context, key K, arg A, f Func[A, V]) (v V, err error, shared bool) {
	c, exists := g.calls.getOrCreate(key)
	defer g.calls.release(c)
	if exists {
		v, err = g.waitCall(ctx, key, c)
		return v, err, true
	}
	return g.doCall(ctx, key, arg, c, f)
}

func (g *Group[K, A, V]) waitCall(ctx context.Context, key K, c *call[V]) (v V, err error) {
	atomic.StoreUint32(&c.shared, 1)
	if g.OnWait != nil {
		g.OnWait(ctx, key)
	}
	done := c.done.get()
	if done != nil {
		select {
		case <-done:
		case <-ctx.Done():
			return v, ctx.Err() //nolint:wrapcheck // Not needed.
		}
	}
	c.checkErrorPanic()
	c.checkErrorGoexit()
	v = c.v
	err = c.err
	return v, err
}

func (g *Group[K, A, V]) doCall(ctx context.Context, key K, arg A, c *call[V], f Func[A, V]) (v V, err error, shared bool) {
	normalReturn := false
	recovered := false
	defer func() {
		if !normalReturn && !recovered {
			c.err = errGoexit
		}
		g.calls.compareAndDelete(key, c)
		c.done.close()
		c.checkErrorPanic()
	}()
	func() {
		defer func() {
			if !normalReturn {
				r := recover()
				if r != nil {
					recovered = true
					c.err = newPanicError(r)
				}
			}
		}()
		c.v, c.err = f(ctx, arg)
		normalReturn = true
	}()
	return c.v, c.err, atomic.LoadUint32(&c.shared) != 0
}

// Forget forgets the call for the given key.
// Future calls to [Group.Do] for this key will call the function rather than waiting for a call to finish.
func (g *Group[K, A, V]) Forget(key K) {
	g.calls.delete(key)
}

type calls[K comparable, V any] struct {
	mu   sync.RWMutex
	m    map[K]*call[V]
	pool syncutil.Pool[*call[V]]
}

func (cs *calls[K, V]) getOrCreate(key K) (c *call[V], exists bool) {
	cs.mu.RLock()
	c, exists = cs.m[key]
	if exists {
		atomic.AddInt32(&c.count, 1)
	}
	cs.mu.RUnlock()
	if exists {
		return c, true
	}
	cs.mu.Lock()
	c, exists = cs.m[key]
	if !exists {
		c = cs.pool.Get()
		if c == nil {
			c = &call[V]{}
		}
		if cs.m == nil {
			cs.m = make(map[K]*call[V])
		}
		cs.m[key] = c
	}
	atomic.AddInt32(&c.count, 1)
	cs.mu.Unlock()
	return c, exists
}

func (cs *calls[K, V]) delete(key K) {
	cs.mu.Lock()
	delete(cs.m, key)
	cs.mu.Unlock()
}

func (cs *calls[K, V]) compareAndDelete(key K, c *call[V]) {
	cs.mu.Lock()
	if cs.m[key] == c {
		delete(cs.m, key)
	}
	cs.mu.Unlock()
}

func (cs *calls[K, V]) release(c *call[V]) {
	count := atomic.AddInt32(&c.count, -1)
	if count == 0 {
		c.reset()
		cs.pool.Put(c)
	}
}

type call[V any] struct {
	done   done
	v      V
	err    error
	shared uint32
	count  int32
}

func (c *call[V]) reset() {
	c.done.reset()
	var zero V
	c.v = zero
	c.err = nil
	c.shared = 0
}

func (c *call[V]) checkErrorPanic() {
	p, ok := c.err.(*panicError) //nolint:errorlint // No need to check the error chain.
	if ok {
		panic(p)
	}
}

func (c *call[V]) checkErrorGoexit() {
	if c.err == errGoexit { //nolint:errorlint // No need to check the error chain.
		runtime.Goexit()
	}
}

type done struct {
	mu          sync.Mutex
	initialized uint32
	ch          chan struct{}
}

func (d *done) reset() {
	d.initialized = 0
	d.ch = nil
}

func (d *done) init(create bool) {
	if atomic.LoadUint32(&d.initialized) == 0 {
		d.mu.Lock()
		if atomic.LoadUint32(&d.initialized) == 0 {
			if create {
				d.ch = make(chan struct{})
			}
			atomic.StoreUint32(&d.initialized, 1)
		}
		d.mu.Unlock()
	}
}

func (d *done) get() <-chan struct{} {
	d.init(true)
	return d.ch
}

func (d *done) close() {
	d.init(false)
	if d.ch != nil {
		close(d.ch)
	}
}

type panicError struct {
	r     any
	stack []byte
}

func newPanicError(r any) error {
	stack := debug.Stack()
	// The first line of the stack trace is of the form "goroutine N [status]:"
	// but by the time the panic reaches Do the goroutine may no longer exist
	// and its status will have changed. Trim out the misleading line.
	if line := bytes.IndexByte(stack, '\n'); line >= 0 {
		stack = stack[line+1:]
	}
	return &panicError{r: r, stack: stack}
}

func (p *panicError) Error() string {
	return fmt.Sprintf("%v\n\n%s", p.r, p.stack)
}

func (p *panicError) Unwrap() error {
	err, _ := p.r.(error)
	return err
}

var errGoexit = errors.New("runtime.Goexit was called")
