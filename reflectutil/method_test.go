//nolint:dupl // Similar but not duplicated.
package reflectutil_test

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/reflectutil"
)

var testMethodCases = []struct {
	name string
	typ  reflect.Type
}{
	{
		name: "String",
		typ:  reflect.TypeFor[string](),
	},
	{
		name: "IOWriter",
		typ:  reflect.TypeFor[io.Writer](),
	},
	{
		name: "BytesBuffer",
		typ:  reflect.TypeFor[*bytes.Buffer](),
	},
}

func TestGetMethods(t *testing.T) {
	for _, tc := range testMethodCases {
		t.Run(tc.name, func(t *testing.T) {
			ms := GetMethods(tc.typ)
			var expected []reflect.Method
			if tc.typ != nil {
				l := tc.typ.NumMethod()
				if l > 0 {
					expected = make([]reflect.Method, l)
					for i := range l {
						expected[i] = tc.typ.Method(i)
					}
				}
			}
			assert.Equal(t, ms.Len(), len(expected))
			for i, m := range ms.All() {
				assert.DeepEqual(t, m, ms.Get(i))
				mn, ok := ms.GetByName(m.Name)
				assert.True(t, ok)
				assert.DeepEqual(t, m, mn)
				assert.DeepEqual(t, m, expected[i])
				if i == ms.Len()-1 {
					break
				}
			}
			assert.AllocsPerRun(t, 100, func() {
				_ = GetMethods(tc.typ)
			}, 0)
		})
	}
}

func BenchmarkGetMethods(b *testing.B) {
	for _, tc := range testMethodCases {
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				_ = GetMethods(tc.typ)
			}
		})
	}
}
