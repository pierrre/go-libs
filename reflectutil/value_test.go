package reflectutil_test

import (
	"io"
	"reflect"
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/reflectutil"
)

func TestValueForInt(t *testing.T) {
	v := ValueFor(3)
	assert.Equal(t, v.Type(), reflect.TypeFor[int]())
}

func TestValueForInterface(t *testing.T) {
	v := ValueFor(io.Writer(nil))
	assert.Equal(t, v.Type(), reflect.TypeFor[io.Writer]())
}

func TestValueForNil(t *testing.T) {
	v := ValueFor[any](nil)
	assert.Equal(t, v.Type(), reflect.TypeFor[any]())
}
