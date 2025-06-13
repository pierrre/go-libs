// Package goroutine helps to manage goroutines safely.
//
// It recovers panic with [panichandle].
package goroutine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
)

// Start executes a function in a new goroutine.
// It returns a function that blocks until the goroutine is terminated.
// The caller must call this function.
func Start(ctx context.Context, f func(ctx context.Context)) (wait func()) {
	var res struct {
		wg     sync.WaitGroup
		panic  error
		goexit bool
	}
	res.wg.Add(1)
	go func() {
		normalReturn := false
		defer func() {
			if !normalReturn {
				r := recover()
				if r != nil {
					res.panic = newPanicError(r)
				} else {
					res.goexit = true
				}
			}
			res.wg.Done()
		}()
		f(ctx)
		normalReturn = true
	}()
	return func() {
		res.wg.Wait()
		if res.goexit {
			runtime.Goexit()
		}
		if res.panic != nil {
			panic(res.panic)
		}
	}
}

// StartWithCancel executes a function in a new goroutine.
// It returns a function that cancels the context and blocks until the goroutine is terminated.
// The caller must call this function.
func StartWithCancel(ctx context.Context, f func(ctx context.Context)) (cancelWait func()) {
	ctx, cancel := context.WithCancel(ctx)
	wait := Start(ctx, f)
	return func() {
		cancel()
		wait()
	}
}

// RunN executes a function with multiple goroutines.
// It blocks until all goroutines are terminated.
func RunN(ctx context.Context, n int, f func(ctx context.Context)) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var res struct {
		wg     sync.WaitGroup
		mu     sync.Mutex
		panics []error
		goexit bool
	}
	res.wg.Add(n)
	for range n {
		go func() {
			runFunc(ctx, f, func(panicErr error, goexit bool) {
				res.mu.Lock()
				if goexit {
					cancel()
					res.goexit = true
				} else if panicErr != nil {
					cancel()
					res.panics = append(res.panics, panicErr)
				}
				res.mu.Unlock()
				res.wg.Done()
			})
		}()
	}
	res.wg.Wait()
	if res.goexit {
		runtime.Goexit()
	}
	if len(res.panics) > 0 {
		err := res.panics[0]
		if len(res.panics) > 1 {
			err = errors.Join(res.panics...)
		}
		panic(err)
	}
}

func runFunc(ctx context.Context, f func(ctx context.Context), after func(panicErr error, goexit bool)) {
	normalReturn := false
	defer func() {
		var panicErr error
		var goexit bool
		if !normalReturn {
			r := recover()
			if r != nil {
				panicErr = newPanicError(r)
			} else {
				goexit = true
			}
		}
		after(panicErr, goexit)
	}()
	f(ctx)
	normalReturn = true
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
