package reflectutil_test

import (
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/pierrre/assert"
	"github.com/pierrre/assert/assertauto"
	. "github.com/pierrre/go-libs/reflectutil"
)

var benchRes any

var types = []reflect.Type{
	reflect.TypeFor[string](),
	reflect.TypeFor[int](),
	reflect.TypeFor[*string](),
	reflect.TypeFor[any](),
	reflect.TypeFor[reflect.Type](),
	reflect.TypeFor[*reflect.Type](),
	reflect.TypeFor[reflect.Value](),
	reflect.TypeFor[*reflect.Value](),
	reflect.TypeFor[atomic.Value](),
	reflect.TypeFor[*atomic.Value](),
}

func TestTypeFullName(t *testing.T) {
	for _, typ := range types {
		s := TypeFullName(typ)
		assertauto.Equal(t, s, assertauto.AssertOptions(assert.MessageWrap(s)))
	}
}

func TestTypeFullNameAllocs(t *testing.T) {
	for _, typ := range types {
		assert.AllocsPerRun(t, 100, func() {
			_ = TypeFullName(typ)
		}, 0, assert.MessageWrap(TypeFullName(typ)))
	}
}

func BenchmarkTypeFullName(b *testing.B) {
	for _, typ := range types {
		b.Run(TypeFullName(typ), func(b *testing.B) {
			var res string
			for range b.N {
				res = TypeFullName(typ)
			}
			benchRes = res
		})
	}
}

func TestTypeFullNameFor(t *testing.T) {
	s := TypeFullNameFor[string]()
	assert.Equal(t, s, "string")
}
