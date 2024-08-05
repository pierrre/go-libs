// Package reflectunsafe provides unsafe operations on reflect package.
package reflectunsafe

import (
	"reflect"
	"unsafe" //nolint:depguard // Required for this package.
)

// ConvertValueCanInterface converts a [reflect.Value] so it can be used with [reflect.Value.Interface].
//
// If [reflect.Value.CanInterface] returns true, it returns the original value.
//
// If the value is not a pointer and [reflect.Value.CanAddr] returns false, it returns the original value.
func ConvertValueCanInterface(v reflect.Value) reflect.Value {
	if v.CanInterface() {
		return v
	}
	typ := v.Type()
	isPtr := false
	var ptr unsafe.Pointer
	switch {
	case typ.Kind() == reflect.Ptr:
		typ = typ.Elem()
		isPtr = true
		ptr = v.UnsafePointer()
	case v.CanAddr():
		ptr = v.Addr().UnsafePointer()
	default:
		return v
	}
	v = reflect.NewAt(typ, ptr)
	if !isPtr {
		v = v.Elem()
	}
	return v
}
