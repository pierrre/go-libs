package reflectutil

import (
	"reflect"
	"strconv"

	"github.com/pierrre/go-libs/syncutil"
)

var typeFullNameCache syncutil.Map[reflect.Type, string]

// TypeFullName returns the full name of the type.
//
// It contains the full package path if the type is defined in a package.
func TypeFullName(typ reflect.Type) string {
	name, ok := typeFullNameCache.Load(typ)
	if ok {
		return name
	}
	name = typeFullName(typ)
	typeFullNameCache.Store(typ, name)
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
		return typ.ChanDir().String() + " " + typeFullName(typ.Elem())
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
