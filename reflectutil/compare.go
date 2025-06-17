package reflectutil

import (
	"bytes"
	"cmp"
	"reflect"
	"strings"

	"github.com/pierrre/go-libs/syncutil"
)

// Compare compares two values of the same [reflect.Type].
func Compare(a, b reflect.Value) int {
	kind := a.Kind()
	if kind == b.Kind() {
		typ := a.Type()
		if typ == b.Type() {
			return GetCompareFunc(typ)(a, b)
		}
	}
	return -1
}

// CompareFunc represents a function that compares 2 [reflect.Value] of the same [reflect.Type].
type CompareFunc func(a, b reflect.Value) int

// GetCompareFunc returns a function that compares two values of the given [reflect.Type].
func GetCompareFunc(typ reflect.Type) CompareFunc {
	if typ == bytesType {
		return compareBytes
	}
	return kindGetCompareFuncs[typ.Kind()](typ)
}

var kindGetCompareFuncs [reflect.UnsafePointer + 1]func(typ reflect.Type) CompareFunc

func init() {
	kindGetCompareFuncs[reflect.Bool] = getCompareFuncBool
	kindGetCompareFuncs[reflect.Int] = getCompareFuncInt
	kindGetCompareFuncs[reflect.Int8] = getCompareFuncInt
	kindGetCompareFuncs[reflect.Int16] = getCompareFuncInt
	kindGetCompareFuncs[reflect.Int32] = getCompareFuncInt
	kindGetCompareFuncs[reflect.Int64] = getCompareFuncInt
	kindGetCompareFuncs[reflect.Uint] = getCompareFuncUint
	kindGetCompareFuncs[reflect.Uint8] = getCompareFuncUint
	kindGetCompareFuncs[reflect.Uint16] = getCompareFuncUint
	kindGetCompareFuncs[reflect.Uint32] = getCompareFuncUint
	kindGetCompareFuncs[reflect.Uint64] = getCompareFuncUint
	kindGetCompareFuncs[reflect.Uintptr] = getCompareFuncUint
	kindGetCompareFuncs[reflect.Float32] = getCompareFuncFloat
	kindGetCompareFuncs[reflect.Float64] = getCompareFuncFloat
	kindGetCompareFuncs[reflect.Complex64] = getCompareFuncComplex
	kindGetCompareFuncs[reflect.Complex128] = getCompareFuncComplex
	kindGetCompareFuncs[reflect.Array] = getCompareFuncArray
	kindGetCompareFuncs[reflect.Chan] = getCompareFuncPointer
	kindGetCompareFuncs[reflect.Func] = getCompareFuncPointer
	kindGetCompareFuncs[reflect.Interface] = getCompareFuncInterface
	kindGetCompareFuncs[reflect.Map] = getCompareFuncPointer
	kindGetCompareFuncs[reflect.Pointer] = getCompareFuncPointer
	kindGetCompareFuncs[reflect.Slice] = getCompareFuncPointer // TODO improve
	kindGetCompareFuncs[reflect.String] = getCompareFuncString
	kindGetCompareFuncs[reflect.Struct] = getCompareFuncStruct
	kindGetCompareFuncs[reflect.UnsafePointer] = getCompareFuncPointer
}

func getCompareFuncBool(typ reflect.Type) CompareFunc {
	return compareBool
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

func getCompareFuncInt(typ reflect.Type) CompareFunc {
	return compareInt
}

func compareInt(a, b reflect.Value) int {
	return cmp.Compare(a.Int(), b.Int())
}

func getCompareFuncUint(typ reflect.Type) CompareFunc {
	return compareUint
}

func compareUint(a, b reflect.Value) int {
	return cmp.Compare(a.Uint(), b.Uint())
}

func getCompareFuncFloat(typ reflect.Type) CompareFunc {
	return compareFloat
}

func compareFloat(a, b reflect.Value) int {
	return cmp.Compare(a.Float(), b.Float())
}

func getCompareFuncComplex(typ reflect.Type) CompareFunc {
	return compareComplex
}

func compareComplex(a, b reflect.Value) int {
	ac := a.Complex()
	bc := b.Complex()
	return cmp.Or(
		cmp.Compare(real(ac), real(bc)),
		cmp.Compare(imag(ac), imag(bc)),
	)
}

func getCompareFuncArray(typ reflect.Type) CompareFunc {
	f, ok := compareFuncArrayCache.Load(typ)
	if ok {
		return f
	}
	elemCmp := GetCompareFunc(typ.Elem())
	f = func(a, b reflect.Value) int {
		for i := range a.Len() {
			c := elemCmp(a.Index(i), b.Index(i))
			if c != 0 {
				return c
			}
		}
		return 0
	}
	compareFuncArrayCache.Store(typ, f)
	return f
}

var compareFuncArrayCache syncutil.Map[reflect.Type, CompareFunc]

func getCompareFuncInterface(typ reflect.Type) CompareFunc {
	return compareInterface
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

func getCompareFuncPointer(typ reflect.Type) CompareFunc {
	return comparePointer
}

func comparePointer(a, b reflect.Value) int {
	return cmp.Compare(uintptr(a.UnsafePointer()), uintptr(b.UnsafePointer()))
}

var bytesType = reflect.TypeFor[[]byte]()

func compareBytes(a, b reflect.Value) int {
	return bytes.Compare(a.Bytes(), b.Bytes())
}

func getCompareFuncString(typ reflect.Type) CompareFunc {
	return compareString
}

func compareString(a, b reflect.Value) int {
	return strings.Compare(a.String(), b.String())
}

func getCompareFuncStruct(typ reflect.Type) CompareFunc {
	f, ok := compareFuncStructCache.Load(typ)
	if ok {
		return f
	}
	fs := GetStructFields(typ)
	l := fs.Len()
	cmps := make([]CompareFunc, l)
	for i := range l {
		cmps[i] = GetCompareFunc(fs.Get(i).Type)
	}
	f = func(a, b reflect.Value) int {
		for i := range l {
			c := cmps[i](a.Field(i), b.Field(i))
			if c != 0 {
				return c
			}
		}
		return 0
	}
	compareFuncStructCache.Store(typ, f)
	return f
}

var compareFuncStructCache syncutil.Map[reflect.Type, CompareFunc]

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
