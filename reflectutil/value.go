package reflectutil

import (
	"reflect"
)

// ValueFor returns a [reflect.Value] of a value, preserving the original type for interfaces.
func ValueFor[T any](i T) reflect.Value {
	return reflect.ValueOf(&i).Elem()
}
