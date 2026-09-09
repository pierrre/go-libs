package weakutil

import (
	"reflect"
)

func IsTypeSafelyComparable(t reflect.Type) bool {
	return isTypeSafelyComparable(t)
}
