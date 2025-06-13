package reflectutil

import (
	"bytes"
	"cmp"
	"reflect"
	"strings"
)

// Compare compares two values of the same [reflect.Type].
func Compare(a, b reflect.Value) int {
	kind := a.Kind()
	if kind == b.Kind() {
		typ := a.Type()
		if typ == b.Type() {
			return getCompareFunc(typ, kind)(a, b)
		}
	}
	return -1
}

// GetCompareFunc returns a function that compares two values of the given [reflect.Type].
//
// The returned function panics if the [reflect.Type] is not supported.
func GetCompareFunc(typ reflect.Type) func(a, b reflect.Value) int {
	return getCompareFunc(typ, typ.Kind())
}

func getCompareFunc(typ reflect.Type, kind reflect.Kind) func(a, b reflect.Value) int {
	if typ == bytesType {
		return compareBytes
	}
	return kindCompareFuncs[kind]
}

var kindCompareFuncs [reflect.UnsafePointer + 1]func(a, b reflect.Value) int

func init() {
	kindCompareFuncs[reflect.Bool] = compareBool
	kindCompareFuncs[reflect.Int] = compareInt
	kindCompareFuncs[reflect.Int8] = compareInt
	kindCompareFuncs[reflect.Int16] = compareInt
	kindCompareFuncs[reflect.Int32] = compareInt
	kindCompareFuncs[reflect.Int64] = compareInt
	kindCompareFuncs[reflect.Uint] = compareUint
	kindCompareFuncs[reflect.Uint8] = compareUint
	kindCompareFuncs[reflect.Uint16] = compareUint
	kindCompareFuncs[reflect.Uint32] = compareUint
	kindCompareFuncs[reflect.Uint64] = compareUint
	kindCompareFuncs[reflect.Uintptr] = compareUint
	kindCompareFuncs[reflect.Float32] = compareFloat
	kindCompareFuncs[reflect.Float64] = compareFloat
	kindCompareFuncs[reflect.Complex64] = compareComplex
	kindCompareFuncs[reflect.Complex128] = compareComplex
	kindCompareFuncs[reflect.Array] = compareArray
	kindCompareFuncs[reflect.Chan] = comparePointer
	kindCompareFuncs[reflect.Func] = comparePointer
	kindCompareFuncs[reflect.Interface] = compareInterface
	kindCompareFuncs[reflect.Map] = comparePointer
	kindCompareFuncs[reflect.Pointer] = comparePointer
	kindCompareFuncs[reflect.Slice] = compareUnsupported // TODO: Implement slice comparison (by pointer).
	kindCompareFuncs[reflect.String] = compareString
	kindCompareFuncs[reflect.Struct] = compareStruct
	kindCompareFuncs[reflect.UnsafePointer] = comparePointer
}

func compareBool(a, b reflect.Value) int {
	ab, bb := a.Bool(), b.Bool()
	switch {
	case ab == bb:
		return 0
	case ab:
		return 1
	default:
		return -1
	}
}

func compareInt(a, b reflect.Value) int {
	return cmp.Compare(a.Int(), b.Int())
}

func compareUint(a, b reflect.Value) int {
	return cmp.Compare(a.Uint(), b.Uint())
}

func compareFloat(a, b reflect.Value) int {
	return cmp.Compare(a.Float(), b.Float())
}

func compareComplex(a, b reflect.Value) int {
	ac := a.Complex()
	bc := b.Complex()
	return cmp.Or(
		cmp.Compare(real(ac), real(bc)),
		cmp.Compare(imag(ac), imag(bc)),
	)
}

func compareArray(a, b reflect.Value) int {
	elemCmp := GetCompareFunc(a.Type().Elem())
	for i := range a.Len() {
		c := elemCmp(a.Index(i), b.Index(i))
		if c != 0 {
			return c
		}
	}
	return 0
}

func compareInterface(a, b reflect.Value) int {
	c, ok := compareNil(a, b)
	if ok {
		return c
	}
	c = Compare(reflect.ValueOf(a.Elem().Type()), reflect.ValueOf(b.Elem().Type()))
	if c != 0 {
		return c
	}
	return Compare(a.Elem(), b.Elem())
}

func comparePointer(a, b reflect.Value) int {
	return cmp.Compare(uintptr(a.UnsafePointer()), uintptr(b.UnsafePointer()))
}

var bytesType = reflect.TypeFor[[]byte]()

func compareBytes(a, b reflect.Value) int {
	return bytes.Compare(a.Bytes(), b.Bytes())
}

func compareString(a, b reflect.Value) int {
	return strings.Compare(a.String(), b.String())
}

func compareStruct(a, b reflect.Value) int {
	for i := range a.NumField() {
		af, bf := a.Field(i), b.Field(i)
		c := getCompareFunc(af.Type(), af.Kind())(af, bf)
		if c != 0 {
			return c
		}
	}
	return 0
}

func compareNil(a, b reflect.Value) (int, bool) {
	if a.IsNil() {
		if b.IsNil() {
			return 0, true
		}
		return -1, true
	}
	if b.IsNil() {
		return 1, true
	}
	return 0, false
}

func compareUnsupported(a, b reflect.Value) int {
	panic("unsupported type: " + a.Type().String())
}
