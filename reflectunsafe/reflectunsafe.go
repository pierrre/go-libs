// Package reflectunsafe provides unsafe operations on reflect package.
package reflectunsafe

import (
	"reflect"
)

// ConvertValueCanInterface converts a [reflect.Value] so it can be used with [reflect.Value.Interface].
//
// If [reflect.Value.CanInterface] returns true, it returns the original value.
//
// If [reflect.Value.CanAddr] returns false, it returns the original value.
func ConvertValueCanInterface(v reflect.Value) reflect.Value {
	if v.CanInterface() {
		return v
	}
	if !v.CanAddr() {
		return v
	}
	ptr := v.Addr().UnsafePointer()
	return reflect.NewAt(v.Type(), ptr).Elem()
}
