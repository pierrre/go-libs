// Package worker provides a way to run a function in a loop, with intervals and error handling.
package worker

import (
	"context"
	"time"
)

// Func represents a function.
type Func func(ctx context.Context)

// Run runs the [Func] in a loop until the [context.Context] is done.
func Run(ctx context.Context, f Func, opts ...Option) {
	o := buildOptions(opts...)
	var ticker *time.Ticker
	if o.interval > 0 {
		ticker = time.NewTicker(o.interval)
		defer ticker.Stop()
	}
	for ctx.Err() == nil {
		if o.immediately {
			f(ctx)
		} else {
			o.immediately = true
		}
		if ticker == nil {
			continue
		}
		if o.fixed {
			tm := time.Now()
			tm = tm.Truncate(o.interval)
			tm = tm.Add(o.interval)
			d := time.Until(tm)
			sleep(ctx, d)
			ticker.Reset(o.interval)
			o.fixed = false
			continue
		}
		waitChan(ctx, ticker.C)
	}
}

type options struct {
	interval    time.Duration
	immediately bool
	fixed       bool
}

func buildOptions(opts ...Option) *options {
	o := &options{
		immediately: true,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// Option is an option for [Run].
type Option func(*options)

// WithInterval sets the interval between each function call.
// The default value is 0, which means the function will be called without any delay between calls.
func WithInterval(d time.Duration) Option {
	return func(o *options) {
		o.interval = d
	}
}

// WithImmediately sets whether to call the function immediately before the first interval.
// The default value is true, which means the function will be called immediately.
func WithImmediately(b bool) Option {
	return func(o *options) {
		o.immediately = b
	}
}

// WithFixed sets whether to call the function at fixed times.
// E.g. if the interval is 1 minute, it will call the function every minute at 0s.
func WithFixed(b bool) Option {
	return func(o *options) {
		o.fixed = b
	}
}

// ErrorFunc represents a function that returns an error.
type ErrorFunc func(ctx context.Context) error

// NewFuncWithError creates a new [Func] that calls the given [ErrorFunc].
// If no error is returned, it will return.
// If an error is returned, it will call the given [OnErrorFunc] if it is not nil.
// If the retry parameter is true, it will retry the [ErrorFunc] until it returns no error or the context is done.
func NewFuncWithError(ef ErrorFunc, onError OnErrorFunc, retry bool) Func {
	return func(ctx context.Context) {
		for ctx.Err() == nil {
			err := ef(ctx)
			if err == nil {
				return
			}
			if onError != nil {
				onError(ctx, err)
			}
			if !retry {
				return
			}
		}
	}
}

// OnErrorFunc represents a function that handles an error.
type OnErrorFunc func(ctx context.Context, err error)

// NewOnErrorFuncWithDelay creates a new [OnErrorFunc] that calls the given [OnErrorFunc], then waits for the given duration.
func NewOnErrorFuncWithDelay(onError OnErrorFunc, d time.Duration) OnErrorFunc {
	return func(ctx context.Context, err error) {
		onError(ctx, err)
		sleep(ctx, d)
	}
}

func sleep(ctx context.Context, d time.Duration) {
	tm := time.NewTimer(d)
	defer tm.Stop()
	waitChan(ctx, tm.C)
}

func waitChan(ctx context.Context, ch <-chan time.Time) {
	select {
	case <-ch:
	case <-ctx.Done():
	}
}
