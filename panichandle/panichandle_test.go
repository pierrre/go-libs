package panichandle

import (
	"testing"
)

func TestDefaultHandler(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("no panic")
		}
	}()
	DefaultHandler("test")
}

func TestRecover(t *testing.T) {
	Recover()
}

func TestRecoverPanic(t *testing.T) {
	defer restoreDefaultHandler()
	called := false
	Handler = func(r interface{}) {
		called = true
		if r == nil {
			t.Fatal("nil")
		}
	}
	defer func() {
		if !called {
			t.Fatal("not called")
		}
	}()
	defer Recover()
	panic("test")
}

func restoreDefaultHandler() {
	Handler = DefaultHandler
}
