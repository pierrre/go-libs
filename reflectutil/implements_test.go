package reflectutil_test

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/pierrre/assert"
	"github.com/pierrre/assert/assertauto"
	. "github.com/pierrre/go-libs/reflectutil"
)

var testImplementsCacheInterfaces = []reflect.Type{
	reflect.TypeFor[any](),
	reflect.TypeFor[io.Writer](),
	reflect.TypeFor[testing.TB](),
	reflect.TypeFor[reflect.Type](),
}

var testImplementsCacheTypes = append([]reflect.Type{
	reflect.TypeFor[string](),
	reflect.TypeFor[*bytes.Buffer](),
	reflect.TypeFor[*testing.T](),
	reflect.TypeOf(reflect.TypeFor[string]()),
}, testImplementsCacheInterfaces...)

func runImplementsCacheTestCases[TB testing.TB](tb TB, f func(tb TB, itf reflect.Type, typ reflect.Type, callImplements func() bool)) {
	for _, itf := range testImplementsCacheInterfaces {
		c := NewImplementsCache(itf)
		for _, typ := range testImplementsCacheTypes {
			callImplements := func() bool {
				return c.ImplementedBy(typ)
			}
			f(tb, itf, typ, callImplements)
		}
	}
}

func TestImplementsCache(t *testing.T) {
	runImplementsCacheTestCases(t, func(t *testing.T, itf reflect.Type, typ reflect.Type, callImplements func() bool) { //nolint:thelper // This is not a test helper.
		t.Run(TypeFullName(itf)+"-"+TypeFullName(typ), func(t *testing.T) {
			implements := callImplements()
			if typ != nil {
				correct := typ.Implements(itf)
				assert.Equal(t, correct, implements)
			} else {
				assert.False(t, implements)
			}
			assert.AllocsPerRun(t, 100, func() {
				_ = callImplements()
			}, 0)
		})
	})
}

func TestNewImplementsCachePanicNil(t *testing.T) {
	rec, _ := assert.Panics(t, func() {
		NewImplementsCache(nil)
	})
	assertauto.Equal(t, rec)
}

func TestNewImplementsCacheForPanicNotInterface(t *testing.T) {
	rec, _ := assert.Panics(t, func() {
		NewImplementsCacheFor[string]()
	})
	assertauto.Equal(t, rec)
}

func BenchmarkImplementsCache(b *testing.B) {
	runImplementsCacheTestCases(b, func(b *testing.B, itf reflect.Type, typ reflect.Type, callImplements func() bool) { //nolint:thelper // This is not a test helper.
		b.Run(TypeFullName(itf)+"-"+TypeFullName(typ), func(b *testing.B) {
			for b.Loop() {
				callImplements()
			}
		})
	})
}
