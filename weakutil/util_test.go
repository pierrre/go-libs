package weakutil_test

import (
	"reflect"
	"runtime/debug"
	"testing"
	"unsafe" //nolint:depguard // needed to test the UnsafePointer kind

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/weakutil"
)

func disableGC(tb testing.TB) {
	tb.Helper()
	old := debug.SetGCPercent(-1)
	tb.Cleanup(func() { debug.SetGCPercent(old) })
}

func TestIsTypeSafelyComparable(t *testing.T) {
	type namedInt int
	type namedStruct struct {
		A int
	}
	type structWithAny struct {
		X any
	}
	type structWithSlice struct {
		S []byte
	}
	type nestedAny struct {
		A structWithAny
	}
	tests := []struct {
		name string
		typ  reflect.Type
		want bool
	}{
		{name: "Bool", typ: reflect.TypeFor[bool](), want: true},
		{name: "Int", typ: reflect.TypeFor[int](), want: true},
		{name: "Int8", typ: reflect.TypeFor[int8](), want: true},
		{name: "Int16", typ: reflect.TypeFor[int16](), want: true},
		{name: "Int32", typ: reflect.TypeFor[int32](), want: true},
		{name: "Int64", typ: reflect.TypeFor[int64](), want: true},
		{name: "Uint", typ: reflect.TypeFor[uint](), want: true},
		{name: "Uint8", typ: reflect.TypeFor[uint8](), want: true},
		{name: "Uint16", typ: reflect.TypeFor[uint16](), want: true},
		{name: "Uint32", typ: reflect.TypeFor[uint32](), want: true},
		{name: "Uint64", typ: reflect.TypeFor[uint64](), want: true},
		{name: "Uintptr", typ: reflect.TypeFor[uintptr](), want: true},
		{name: "Float32", typ: reflect.TypeFor[float32](), want: true},
		{name: "Float64", typ: reflect.TypeFor[float64](), want: true},
		{name: "Complex64", typ: reflect.TypeFor[complex64](), want: true},
		{name: "Complex128", typ: reflect.TypeFor[complex128](), want: true},
		{name: "String", typ: reflect.TypeFor[string](), want: true},
		{name: "Ptr", typ: reflect.TypeFor[*int](), want: true},
		{name: "Chan", typ: reflect.TypeFor[chan int](), want: true},
		{name: "UnsafePointer", typ: reflect.TypeFor[unsafe.Pointer](), want: true},
		{name: "ArrayOfInt", typ: reflect.TypeFor[[3]int](), want: true},
		{name: "ArrayOfIntLen0", typ: reflect.TypeFor[[0]int](), want: true},
		{name: "EmptyStruct", typ: reflect.TypeFor[struct{}](), want: true},
		{name: "PlainStruct", typ: reflect.TypeFor[struct {
			A int
			B string
		}](), want: true},
		{name: "StructWithPtrAndChan", typ: reflect.TypeFor[struct {
			P *int
			C chan int
		}](), want: true},
		{name: "NamedInt", typ: reflect.TypeFor[namedInt](), want: true},
		{name: "NamedStruct", typ: reflect.TypeFor[namedStruct](), want: true},
		{name: "Any", typ: reflect.TypeFor[any](), want: false},
		{name: "InterfaceWithMethod", typ: reflect.TypeFor[interface{ M() }](), want: false},
		{name: "Slice", typ: reflect.TypeFor[[]int](), want: false},
		{name: "Map", typ: reflect.TypeFor[map[string]int](), want: false},
		{name: "Func", typ: reflect.TypeFor[func()](), want: false},
		{name: "StructWithAny", typ: reflect.TypeFor[structWithAny](), want: false},
		{name: "ArrayOfAny", typ: reflect.TypeFor[[1]any](), want: false},
		{name: "StructWithSlice", typ: reflect.TypeFor[structWithSlice](), want: false},
		{name: "NestedStructWithAny", typ: reflect.TypeFor[nestedAny](), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, IsTypeSafelyComparable(tt.typ), tt.want)
		})
	}
}
