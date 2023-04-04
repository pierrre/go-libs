package panichandle

import (
	"testing"

	"github.com/pierrre/assert"
	"github.com/pierrre/go-libs/internal/golibstest"
)

func init() {
	golibstest.Configure()
}

func TestRecoverNoPanicWithoutHandler(t *testing.T) {
	Recover()
}

func TestRecoverNoPanicWithandler(t *testing.T) {
	defer func() {
		Handler = nil
	}()
	called := false
	Handler = func(r any) {
		called = true
	}
	defer func() {
		assert.False(t, called)
	}()
	Recover()
}

func TestRecoverPanicWithoutHandler(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotZero(t, r)
	}()
	defer Recover()
	panic("test")
}

func TestRecoverPanicWithHandler(t *testing.T) {
	defer func() {
		Handler = nil
	}()
	called := false
	Handler = func(r any) {
		called = true
		assert.NotZero(t, r)
	}
	defer func() {
		assert.True(t, called)
	}()
	defer Recover()
	panic("test")
}
