package reflectutil

import (
	"reflect"

	"github.com/pierrre/go-libs/syncutil"
)

// IsTypeStrictlyComparable reports whether values of type t can always be compared with == without panicking at runtime.
// It is conservative: it returns false for any type it cannot prove safe.
func IsTypeStrictlyComparable(t reflect.Type) bool {
	switch t.Kind() { //nolint:exhaustive // the default case handles all the remaining kinds
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.String,
		reflect.Pointer, reflect.Chan, reflect.UnsafePointer:
		return true
	case reflect.Array:
		return IsTypeStrictlyComparable(t.Elem())
	case reflect.Struct:
		return isStructStrictlyComparable(t)
	default:
		// Interface (any), Slice, Map, Func, Invalid: not provably safe.
		return false
	}
}

var comparableTypeCache syncutil.Map[reflect.Type, bool]

func isStructStrictlyComparable(t reflect.Type) bool {
	if v, ok := comparableTypeCache.Load(t); ok {
		return v
	}
	fs := GetStructFields(t)
	result := true
	for i := range fs.Len() {
		if !IsTypeStrictlyComparable(fs.Get(i).Type) {
			result = false
			break
		}
	}
	comparableTypeCache.LoadOrStore(t, result)
	return result
}
