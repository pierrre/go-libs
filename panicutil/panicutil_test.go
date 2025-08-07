package panicutil_test

import (
	"errors"
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/panicutil"
)

func TestNewError(t *testing.T) {
	r := errors.New("error")
	err := NewError(r)
	assert.Error(t, err)
	s := err.Error()
	assert.NotZero(t, s)
	t.Log(s)
	assert.Error(t, errors.Unwrap(err))
	esf, _ := assert.Type[interface{ StackFrames() []uintptr }](t, err)
	assert.SliceNotEmpty(t, esf.StackFrames())
}

func BenchmarkError(b *testing.B) {
	r := errors.New("error")
	err := NewError(r)
	for b.Loop() {
		_ = err.Error()
	}
}
