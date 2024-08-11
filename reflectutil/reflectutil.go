// Package reflectutil provides utility functions for the reflect package.
package reflectutil

import (
	"reflect"
	"strconv"
	"sync"
)

var (
	typeFullNameCache   = make(map[reflect.Type]string)
	typeFullNameCacheMu sync.Mutex
)

// TypeFullName returns the full name of the type.
//
// It contains the full package path if the type is defined in a package.
func TypeFullName(typ reflect.Type) string {
	typeFullNameCacheMu.Lock()
	defer typeFullNameCacheMu.Unlock()
	name, ok := typeFullNameCache[typ]
	if !ok {
		name = typeFullName(typ)
		typeFullNameCache[typ] = name
	}
	return name
}

func typeFullName(typ reflect.Type) string {
	pkgPath := typ.PkgPath()
	if pkgPath != "" {
		return pkgPath + "." + typ.Name()
	}
	switch typ.Kind() { //nolint:exhaustive // We only need to handle composite types.
	case reflect.Pointer:
		return "*" + typeFullName(typ.Elem())
	case reflect.Slice:
		return "[]" + typeFullName(typ.Elem())
	case reflect.Array:
		return "[" + strconv.Itoa(typ.Len()) + "]" + typeFullName(typ.Elem())
	case reflect.Chan:
		return "chan " + typeFullName(typ.Elem())
	case reflect.Map:
		return "map[" + typeFullName(typ.Key()) + "]" + typeFullName(typ.Elem())
	}
	return typ.String()
}

func TypeFullNameFor[T any]() string {
	s := reflect.TypeFor[typeContainer[T]]().String()
	return s[typeContainerPrefixLen : len(s)-typeContainerSuffixLen]
}

type typeContainer[T any] struct{}

var (
	typeContainerSuffixLen = len("]")
	typeContainerPrefixLen = len(reflect.TypeFor[typeContainer[string]]().String()) - len("string") - typeContainerSuffixLen
)

// ValueInterfaceFor calls [reflect.Value.Interface] and returns the result as the specified type.
//
// The caller must ensure that the type is correct.
func ValueInterfaceFor[T any](v reflect.Value) T {
	return v.Interface().(T) //nolint:forcetypeassert // It should be checked by the caller.
}

// ConvertValueCanInterface attempts to converts a [reflect.Value] so it can be used with [reflect.Value.Interface].
//
// The returned boolean indicates if the conversion was successful.
//
// If the conversion was successful, the returned [reflect.Value] can be used with [reflect.Value.Interface].
// If the conversion was not successful, the returned [reflect.Value] is the same as the input [reflect.Value].
func ConvertValueCanInterface(v reflect.Value) (reflect.Value, bool) {
	if v.CanInterface() {
		return v, true
	}
	if v.Kind() == reflect.Ptr {
		return reflect.NewAt(v.Type().Elem(), v.UnsafePointer()), true
	}
	if v.CanAddr() {
		return reflect.NewAt(v.Type(), v.Addr().UnsafePointer()).Elem(), true
	}
	return v, false
}

// TryValueInterface calls [ConvertValueCanInterface] and [reflect.Value.Interface].
func TryValueInterface(v reflect.Value) (any, bool) {
	v, ok := ConvertValueCanInterface(v)
	if !ok {
		return nil, false
	}
	return v.Interface(), true
}

// TryValueInterfaceFor calls [ConvertValueCanInterface] and [ValueInterfaceFor].
func TryValueInterfaceFor[T any](v reflect.Value) (T, bool) {
	v, ok := ConvertValueCanInterface(v)
	if !ok {
		var zero T
		return zero, false
	}
	return ValueInterfaceFor[T](v), true
}
