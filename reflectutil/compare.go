package reflectutil

import (
	"cmp"
	"reflect"
)

// Compare compares two values of the same type.
func Compare(a, b reflect.Value) int {
	if a.Type() != b.Type() {
		return -1
	}
	return GetCompareFunc(a.Type())(a, b)
}

// GetCompareFunc returns a function that compares two values of the given type.
//
// It panics if the type is not supported.
//
//nolint:gocyclo // We need to handle all types.
func GetCompareFunc(typ reflect.Type) func(a, b reflect.Value) int {
	switch typ.Kind() { //nolint:exhaustive // Optimized for common kinds, the default case is less optimized.
	case reflect.Bool:
		return compareBool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return compareInt
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return compareUint
	case reflect.Float32, reflect.Float64:
		return compareFloat
	case reflect.Complex64, reflect.Complex128:
		return compareComplex
	case reflect.String:
		return compareString
	case reflect.Pointer, reflect.UnsafePointer:
		return comparePointer
	case reflect.Chan:
		return compareChan
	case reflect.Array:
		return compareArray
	case reflect.Struct:
		return compareStruct
	case reflect.Interface:
		return compareInterface
	}
	panic("unsupported type: " + typ.String())
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

func compareString(a, b reflect.Value) int {
	return cmp.Compare(a.String(), b.String())
}

func comparePointer(a, b reflect.Value) int {
	return cmp.Compare(a.Pointer(), b.Pointer())
}

func compareChan(a, b reflect.Value) int {
	c, ok := compareNil(a, b)
	if ok {
		return c
	}
	return cmp.Compare(a.Pointer(), b.Pointer())
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

func compareStruct(a, b reflect.Value) int {
	for i := range a.NumField() {
		c := Compare(a.Field(i), b.Field(i))
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
