//nolint:dupl // Similar but not duplicated.
package reflectutil_test

import (
	"reflect"
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/reflectutil"
)

var testStructFieldTypes = []struct {
	name string
	typ  reflect.Type
}{
	{
		name: "Nil",
		typ:  nil,
	},
	{
		name: "EmptyStruct",
		typ:  reflect.TypeFor[struct{}](),
	},
	{
		name: "CustomStruct",
		typ: func() reflect.Type {
			type CustomStruct struct {
				String string
				Int    int
				Float  float64
				Bool   bool
			}
			return reflect.TypeFor[CustomStruct]()
		}(),
	},
}

func TestGetStructFields(t *testing.T) {
	for _, tc := range testStructFieldTypes {
		t.Run(tc.name, func(t *testing.T) {
			fs := GetStructFields(tc.typ)
			var expected []reflect.StructField
			if tc.typ != nil {
				l := tc.typ.NumField()
				if l > 0 {
					expected = make([]reflect.StructField, l)
					for i := range l {
						expected[i] = tc.typ.Field(i)
					}
				}
			}
			assert.Equal(t, fs.Len(), len(expected))
			for i, f := range fs.All() {
				assert.DeepEqual(t, f, fs.Get(i))
				assert.DeepEqual(t, f, expected[i])
				if i == fs.Len()-1 {
					break
				}
			}
			assert.AllocsPerRun(t, 100, func() {
				_ = GetStructFields(tc.typ)
			}, 0)
		})
	}
}

func BenchmarkGetStructFields(b *testing.B) {
	for _, tc := range testStructFieldTypes {
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				_ = GetStructFields(tc.typ)
			}
		})
	}
}
