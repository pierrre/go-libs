package singleflight_test

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/singleflight"
)

func TestDo(t *testing.T) {
	ctx := t.Context()
	g := &Group[string, string, int]{}
	var called atomic.Uint32
	v, err, shared := g.Do(ctx, "key", "arg", func(_ context.Context, arg string) (int, error) {
		called.Add(1)
		assert.Equal(t, arg, "arg")
		return 123, nil
	})
	assert.NoError(t, err)
	assert.False(t, shared)
	assert.Equal(t, v, 123)
	assert.Equal(t, called.Load(), 1)
}

func TestDoError(t *testing.T) {
	ctx := t.Context()
	g := &Group[string, string, int]{}
	var called atomic.Uint32
	v, err, shared := g.Do(ctx, "key", "arg", func(_ context.Context, _ string) (int, error) {
		called.Add(1)
		return 123, errors.New("error")
	})
	assert.Error(t, err)
	assert.False(t, shared)
	assert.Equal(t, v, 123)
	assert.Equal(t, called.Load(), 1)
}

func TestDoContextCancel(t *testing.T) {
	ctx := t.Context()
	ctx, cancel := context.WithCancel(ctx)
	g := &Group[string, string, int]{}
	var called atomic.Uint32
	v, err, shared := g.Do(ctx, "key", "arg", func(ctx context.Context, _ string) (int, error) {
		called.Add(1)
		cancel()
		return 123, ctx.Err()
	})
	assert.ErrorIs(t, err, context.Canceled)
	assert.False(t, shared)
	assert.Equal(t, v, 123)
	assert.Equal(t, called.Load(), 1)
}

func TestDoPanic(t *testing.T) {
	ctx := t.Context()
	g := &Group[string, string, int]{}
	var called atomic.Uint32
	r, _ := assert.Panics(t, func() {
		_, _, _ = g.Do(ctx, "key", "arg", func(_ context.Context, _ string) (int, error) {
			called.Add(1)
			panic(errors.New("error"))
		})
	})
	err, _ := assert.Type[error](t, r)
	assert.Error(t, err)
	t.Log(err)
	err = errors.Unwrap(err)
	assert.Error(t, err)
	assert.Equal(t, called.Load(), 1)
}

func TestDoGoexit(t *testing.T) {
	ctx := t.Context()
	g := &Group[string, string, int]{}
	normalReturn := false
	recovered := false
	var v int
	var err error
	var shared bool
	done := make(chan struct{})
	var called atomic.Uint32
	go func() {
		defer func() {
			r := recover()
			if r != nil {
				recovered = true
			}
			close(done)
		}()
		v, err, shared = g.Do(ctx, "key", "arg", func(_ context.Context, _ string) (int, error) {
			called.Add(1)
			runtime.Goexit()
			panic("should not be called")
		})
		normalReturn = true
	}()
	<-done
	assert.False(t, normalReturn)
	assert.False(t, recovered)
	assert.NoError(t, err)
	assert.False(t, shared)
	assert.Equal(t, v, 0)
	assert.Equal(t, called.Load(), 1)
}

func TestDoSharedCall(t *testing.T) {
	ctx := t.Context()
	calling := make(chan struct{})
	waiting := make(chan struct{})
	g := &Group[string, string, int]{
		OnWait: func(ctx context.Context, key string) {
			close(waiting)
		},
	}
	var called atomic.Uint32
	go func() {
		<-calling
		_, _, _ = g.Do(ctx, "key", "other", func(_ context.Context, _ string) (int, error) {
			called.Add(1)
			panic("should not be called")
		})
	}()
	v, err, shared := g.Do(ctx, "key", "arg", func(_ context.Context, arg string) (int, error) {
		called.Add(1)
		close(calling)
		assert.Equal(t, arg, arg)
		<-waiting
		return 123, nil
	})
	assert.NoError(t, err)
	assert.True(t, shared)
	assert.Equal(t, v, 123)
	assert.Equal(t, called.Load(), 1)
}

func TestDoWait(t *testing.T) {
	ctx := t.Context()
	calling := make(chan struct{})
	waiting := make(chan struct{})
	g := &Group[string, string, int]{
		OnWait: func(ctx context.Context, key string) {
			close(waiting)
		},
	}
	var called atomic.Uint32
	go func() {
		_, _, _ = g.Do(ctx, "key", "arg", func(_ context.Context, arg string) (int, error) {
			called.Add(1)
			close(calling)
			assert.Equal(t, arg, "arg")
			<-waiting
			return 123, nil
		})
	}()
	<-calling
	v, err, shared := g.Do(ctx, "key", "other", func(_ context.Context, _ string) (int, error) {
		called.Add(1)
		panic("should not be called")
	})
	assert.NoError(t, err)
	assert.True(t, shared)
	assert.Equal(t, v, 123)
	assert.Equal(t, called.Load(), 1)
}

func TestDoWaitContextCancel(t *testing.T) {
	ctx := t.Context()
	calling := make(chan struct{})
	waiting := make(chan struct{})
	g := &Group[string, string, int]{
		OnWait: func(ctx context.Context, key string) {
			close(waiting)
		},
	}
	var called atomic.Uint32
	go func() {
		_, _, _ = g.Do(ctx, "key", "arg", func(_ context.Context, _ string) (int, error) {
			called.Add(1)
			close(calling)
			<-waiting
			return 123, nil
		})
	}()
	<-calling
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	v, err, shared := g.Do(ctx, "key", "other", func(_ context.Context, _ string) (int, error) {
		called.Add(1)
		panic("should not be called")
	})
	assert.ErrorIs(t, err, context.Canceled)
	assert.True(t, shared)
	assert.Equal(t, v, 0)
	assert.Equal(t, called.Load(), 1)
}

func TestDoWaitPanic(t *testing.T) {
	ctx := t.Context()
	calling := make(chan struct{})
	waiting := make(chan struct{})
	g := &Group[string, string, int]{
		OnWait: func(ctx context.Context, key string) {
			close(waiting)
		},
	}
	var called atomic.Uint32
	go func() {
		defer func() {
			_ = recover()
		}()
		_, _, _ = g.Do(ctx, "key", "arg", func(_ context.Context, _ string) (int, error) {
			called.Add(1)
			close(calling)
			<-waiting
			panic("panic")
		})
	}()
	<-calling
	r, _ := assert.Panics(t, func() {
		_, _, _ = g.Do(ctx, "key", "other", func(_ context.Context, _ string) (int, error) {
			called.Add(1)
			return 123, errors.New("should not be called")
		})
	})
	err, _ := assert.Type[error](t, r)
	assert.Error(t, err)
	t.Log(err)
	assert.Equal(t, called.Load(), 1)
}

func TestDoWaitGoexit(t *testing.T) {
	ctx := t.Context()
	calling := make(chan struct{})
	waiting := make(chan struct{})
	g := &Group[string, string, int]{
		OnWait: func(ctx context.Context, key string) {
			close(waiting)
		},
	}
	var called atomic.Uint32
	go func() {
		_, _, _ = g.Do(ctx, "key", "arg", func(_ context.Context, _ string) (int, error) {
			called.Add(1)
			close(calling)
			<-waiting
			runtime.Goexit()
			panic("should not be called")
		})
	}()
	<-calling
	normalReturn := false
	recovered := false
	var v int
	var err error
	var shared bool
	done := make(chan struct{})
	go func() {
		defer func() {
			r := recover()
			if r != nil {
				recovered = true
			}
			close(done)
		}()
		v, err, shared = g.Do(ctx, "key", "other", func(_ context.Context, _ string) (int, error) {
			called.Add(1)
			panic("should not be called")
		})
		normalReturn = true
	}()
	<-done
	assert.False(t, normalReturn)
	assert.False(t, recovered)
	assert.NoError(t, err)
	assert.False(t, shared)
	assert.Equal(t, v, 0)
	assert.Equal(t, called.Load(), 1)
}

func TestConcurrency(t *testing.T) {
	ctx := t.Context()
	g := &Group[int, int, int]{}
	wg := new(sync.WaitGroup)
	for range 10 {
		for range 10 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := range 1000 {
					key := i % 10
					v, err, _ := g.Do(ctx, key, key, func(ctx context.Context, arg int) (int, error) {
						assert.Equal(t, arg, key, assert.Report(testing.TB.Error))
						runtime.Gosched()
						return arg, nil
					})
					assert.NoError(t, err, assert.Report(testing.TB.Error))
					assert.Equal(t, v, key, assert.Report(testing.TB.Error))
				}
			}()
		}
		wg.Wait()
	}
}

func TestForget(t *testing.T) {
	ctx := t.Context()
	calling1 := make(chan struct{})
	calling2 := make(chan struct{})
	g := &Group[string, string, int]{}
	var called atomic.Uint32
	go func() {
		_, _, _ = g.Do(ctx, "key", "arg1", func(_ context.Context, arg string) (int, error) {
			called.Add(1)
			close(calling1)
			<-calling2
			return 123, nil
		})
	}()
	<-calling1
	g.Forget("key")
	v, err, shared := g.Do(ctx, "key", "arg2", func(_ context.Context, arg string) (int, error) {
		called.Add(1)
		close(calling2)
		assert.Equal(t, arg, "arg2")
		return 456, nil
	})
	assert.NoError(t, err)
	assert.False(t, shared)
	assert.Equal(t, v, 456)
	assert.Equal(t, called.Load(), 2)
}

func TestForgetNotFound(t *testing.T) {
	g := &Group[string, string, int]{}
	g.Forget("key")
}

func BenchmarkDo(b *testing.B) {
	ctx := b.Context()
	g := &Group[string, string, int]{}
	key := "key"
	arg := "arg"
	f := func(_ context.Context, _ string) (int, error) {
		return 123, nil
	}
	for b.Loop() {
		_, _, _ = g.Do(ctx, key, arg, f)
	}
}

func BenchmarkDoParallel(b *testing.B) {
	ctx := b.Context()
	g := &Group[string, string, int]{}
	key := "key"
	arg := "arg"
	f := func(_ context.Context, _ string) (int, error) {
		return 123, nil
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _, _ = g.Do(ctx, key, arg, f)
		}
	})
}
