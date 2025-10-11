package worker

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"testing/synctest"
	"time"

	"github.com/pierrre/assert"
)

func Example() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	called := 0
	ef := func(ctx context.Context) error {
		called++
		fmt.Println("call:", called)
		if called <= 5 {
			return errors.New("error")
		}
		if called >= 10 {
			cancel()
		}
		return nil
	}

	onError := func(ctx context.Context, err error) {
		fmt.Println("on error:", err)
	}
	onError = NewOnErrorFuncWithDelay(onError, 1*time.Microsecond)

	f := NewFuncWithError(ef, onError, true)

	Run(ctx, f, WithImmediately(true), WithInterval(10*time.Microsecond), WithFixed(true))

	// Output:
	// call: 1
	// on error: error
	// call: 2
	// on error: error
	// call: 3
	// on error: error
	// call: 4
	// on error: error
	// call: 5
	// on error: error
	// call: 6
	// call: 7
	// call: 8
	// call: 9
	// call: 10
}

func TestRun(t *testing.T) {
	ctx := t.Context()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	called := 0
	f := func(ctx context.Context) {
		called++
		if called == 10 {
			cancel()
		}
	}
	Run(ctx, f)
	assert.Equal(t, called, 10)
}

func TestRunWithInterval(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		called := 0
		var previous time.Time
		f := func(ctx context.Context) {
			called++
			if called == 10 {
				cancel()
			}
			now := time.Now()
			if called > 1 {
				assert.Equal(t, now.Sub(previous), 1*time.Second)
			}
			previous = now
		}
		Run(ctx, f, WithInterval(1*time.Second))
		assert.Equal(t, called, 10)
	})
}

func TestRunWithImmediatelyFalse(t *testing.T) {
	ctx := t.Context()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	called := 0
	f := func(ctx context.Context) {
		called++
		if called == 10 {
			cancel()
		}
	}
	Run(ctx, f, WithImmediately(false))
	assert.Equal(t, called, 10)
}

func TestRunWithFixed(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		called := 0
		f := func(ctx context.Context) {
			called++
			if called == 10 {
				cancel()
			}
			if called > 1 {
				assert.Equal(t, time.Now().Nanosecond(), 0)
			}
		}
		time.Sleep(500 * time.Millisecond)
		Run(ctx, f, WithInterval(1*time.Second), WithFixed(true))
		assert.Equal(t, called, 10)
	})
}

func TestNewFuncWithError(t *testing.T) {
	ctx := t.Context()
	called := 0
	ef := func(ctx context.Context) error {
		called++
		if called < 10 {
			return errors.New("error")
		}
		return nil
	}
	onErrorCalled := 0
	onError := func(ctx context.Context, err error) {
		onErrorCalled++
	}
	f := NewFuncWithError(ef, onError, true)
	f(ctx)
	assert.Equal(t, onErrorCalled, 9)
}

func TestNewFuncWithErrorNoRetry(t *testing.T) {
	ctx := t.Context()
	called := 0
	ef := func(ctx context.Context) error {
		called++
		return errors.New("error")
	}
	f := NewFuncWithError(ef, nil, false)
	f(ctx)
	assert.Equal(t, called, 1)
}

func TestNewOnErrorFuncWithDelay(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		called := 0
		onError := func(ctx context.Context, err error) {
			called++
		}
		start := time.Now()
		onError = NewOnErrorFuncWithDelay(onError, 1*time.Second)
		onError(ctx, errors.New("error"))
		assert.Equal(t, called, 1)
		assert.Equal(t, time.Since(start), 1*time.Second)
	})
}
