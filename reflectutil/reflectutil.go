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
		pkgPath := typ.PkgPath()
		if pkgPath != "" {
			name = pkgPath + "." + typ.Name()
		} else {
			name = typ.String()
		}
		typeFullNameCache[typ] = name
	}
	return name
}

// TypeFullNameFor returns the full name of the argument type.
//
// See [TypeFullName] for more information.
func TypeFullNameFor[T any]() string {
	return TypeFullName(reflect.TypeFor[T]())
}
