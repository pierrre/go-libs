package reflectutil_test

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/reflectutil"
)

var testMethodTypes = []reflect.Type{
	nil,
	reflect.TypeFor[string](),
	reflect.TypeFor[io.Writer](),
	reflect.TypeFor[*bytes.Buffer](),
}

func TestGetMethods(t *testing.T) {
	for _, typ := range testMethodTypes {
		t.Run(TypeFullName(typ), func(t *testing.T) {
			ms := GetMethods(typ)
			var expected []reflect.Method
			if typ != nil {
				l := typ.NumMethod()
				if l > 0 {
					expected = make([]reflect.Method, l)
					for i := range l {
						expected[i] = typ.Method(i)
					}
				}
			}
			assert.Equal(t, ms.Len(), len(expected))
			for i, m := range ms.All() {
				assert.Equal(t, m, ms.Get(i))
				assert.Equal(t, m, expected[i])
				if i == ms.Len()-1 {
					break
				}
			}
			assert.AllocsPerRun(t, 100, func() {
				_ = GetMethods(typ)
			}, 0)
		})
	}
}

func BenchmarkGetMethods(b *testing.B) {
	for _, typ := range testMethodTypes {
		b.Run(TypeFullName(typ), func(b *testing.B) {
			for b.Loop() {
				_ = GetMethods(typ)
			}
		})
	}
}
