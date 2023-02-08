package panichandle

import (
	"testing"

	"github.com/pierrre/assert"
	"github.com/pierrre/go-libs/internal/golibstest"
)

func init() {
	golibstest.Configure()
}

func TestDefaultHandler(t *testing.T) {
	assert.Panics(t, func() {
		DefaultHandler("test")
	})
}

func TestRecover(t *testing.T) {
	Recover()
}

func TestRecoverPanic(t *testing.T) {
	defer restoreDefaultHandler()
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

func restoreDefaultHandler() {
	Handler = DefaultHandler
}
