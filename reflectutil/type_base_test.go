package reflectutil_test

import (
	"reflect"
	"testing"
	"unsafe" //nolint:depguard // Required for unsafe.Pointer.

	"github.com/pierrre/assert"
	"github.com/pierrre/assert/assertauto"
	. "github.com/pierrre/go-libs/reflectutil"
)

var testBaseTypes = []reflect.Type{
	nil,
	reflect.TypeFor[bool](),
	reflect.TypeFor[int](),
	reflect.TypeFor[int8](),
	reflect.TypeFor[int16](),
	reflect.TypeFor[int32](),
	reflect.TypeFor[int64](),
	reflect.TypeFor[uint](),
	reflect.TypeFor[uint8](),
	reflect.TypeFor[uint16](),
	reflect.TypeFor[uint32](),
	reflect.TypeFor[uint64](),
	reflect.TypeFor[uintptr](),
	reflect.TypeFor[float32](),
	reflect.TypeFor[float64](),
	reflect.TypeFor[complex64](),
	reflect.TypeFor[complex128](),
	reflect.TypeFor[[1]int](),
	reflect.TypeFor[chan int](),
	reflect.TypeFor[func(int) int](),
	reflect.TypeFor[map[string]int](),
	reflect.TypeFor[*string](),
	reflect.TypeFor[[]int](),
	reflect.TypeFor[string](),
	reflect.TypeFor[unsafe.Pointer](),
	reflect.TypeFor[testing.TB](),
	reflect.TypeFor[struct{ String string }](),
	func() reflect.Type {
		type CustomString string
		return reflect.TypeFor[CustomString]()
	}(),
	func() reflect.Type {
		type CustomArray [1]int
		return reflect.TypeFor[CustomArray]()
	}(),
	func() reflect.Type {
		type CustomChan chan int
		return reflect.TypeFor[CustomChan]()
	}(),
	func() reflect.Type {
		type CustomFunc func(int) int
		return reflect.TypeFor[CustomFunc]()
	}(),
	func() reflect.Type {
		type CustomMap map[string]int
		return reflect.TypeFor[CustomMap]()
	}(),
	func() reflect.Type {
		type CustomPointer *string
		return reflect.TypeFor[CustomPointer]()
	}(),
	func() reflect.Type {
		type CustomSlice []int
		return reflect.TypeFor[CustomSlice]()
	}(),
	func() reflect.Type {
		type CustomString string
		return reflect.TypeFor[CustomString]()
	}(),
	func() reflect.Type {
		type CustomStruct struct {
			String string
		}
		return reflect.TypeFor[CustomStruct]()
	}(),
}

func TestGetBaseType(t *testing.T) {
	type result struct {
		typ  string
		base string
	}
	for _, typ := range testBaseTypes {
		base := GetBaseType(typ)
		res := result{
			typ:  TypeFullName(typ),
			base: TypeFullName(base),
		}
		assertauto.Equal(t, res)
		assert.AllocsPerRun(t, 100, func() {
			_ = GetBaseType(typ)
		}, 0)
	}
}

func BenchmarkGetBaseType(b *testing.B) {
	for _, typ := range testBaseTypes {
		b.Run(TypeFullName(typ), func(b *testing.B) {
			for b.Loop() {
				_ = GetBaseType(typ)
			}
		})
	}
}
