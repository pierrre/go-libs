package funcutil_test

import (
	"errors"
	"runtime"
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/funcutil"
)

func TestCall(t *testing.T) {
	fCalled := false
	afterCalled := false
	Call(
		func() {
			fCalled = true
		},
		func(goexit bool, panicErr error) {
			afterCalled = true
			assert.True(t, fCalled)
			assert.False(t, goexit)
			assert.NoError(t, panicErr)
		},
	)
	assert.True(t, afterCalled)
}

func TestCallPanic(t *testing.T) {
	var afterGoexit bool
	var afterPanicErr error
	done := make(chan struct{})
	go Call(
		func() {
			panic(errors.New("error"))
		},
		func(goexit bool, panicErr error) {
			afterGoexit = goexit
			afterPanicErr = panicErr
			close(done)
		},
	)
	<-done
	assert.False(t, afterGoexit)
	assert.Error(t, afterPanicErr)
	t.Log(afterPanicErr)
	err := errors.Unwrap(afterPanicErr)
	assert.Error(t, err)
}

func TestCallGoexit(t *testing.T) {
	var afterGoexit bool
	var afterPanicErr error
	done := make(chan struct{})
	go Call(
		func() {
			runtime.Goexit()
		},
		func(goexit bool, panicErr error) {
			afterGoexit = goexit
			afterPanicErr = panicErr
			close(done)
		},
	)
	<-done
	assert.True(t, afterGoexit)
	assert.NoError(t, afterPanicErr)
}

func TestCallPanicAndGoexit(t *testing.T) {
	var afterGoexit bool
	var afterPanicErr error
	done := make(chan struct{})
	go Call(
		func() {
			defer runtime.Goexit()
			panic(errors.New("error"))
		},
		func(goexit bool, panicErr error) {
			afterGoexit = goexit
			afterPanicErr = panicErr
			close(done)
		},
	)
	<-done
	assert.True(t, afterGoexit)
	assert.NoError(t, afterPanicErr)
}

func TestCallGoexitAndPanic(t *testing.T) {
	var afterGoexit bool
	var afterPanicErr error
	done := make(chan struct{})
	go Call(
		func() {
			defer panic(errors.New("error"))
			runtime.Goexit()
		},
		func(goexit bool, panicErr error) {
			afterGoexit = goexit
			afterPanicErr = panicErr
			close(done)
		},
	)
	<-done
	assert.True(t, afterGoexit)
	assert.Error(t, afterPanicErr)
}

func BenchmarkCall(b *testing.B) {
	for b.Loop() {
		Call(
			func() {},
			func(goexit bool, panicErr error) {},
		)
	}
}
