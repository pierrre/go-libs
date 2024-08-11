// Package reflectutil provides utility functions for the reflect package.
package reflectutil

import (
	"reflect"
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
	if typ.Kind() == reflect.Ptr {
		return "*" + typeFullName(typ.Elem())
	}
	pkgPath := typ.PkgPath()
	if pkgPath != "" {
		return pkgPath + "." + typ.Name()
	}
	return typ.String()
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
