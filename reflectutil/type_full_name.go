package reflectutil

import (
	"reflect"
	"strconv"
	"unsafe" //nolint:depguard // Required for unsafe.Pointer.

	"github.com/pierrre/go-libs/syncutil"
)

var typeFullNameCache syncutil.Map[reflect.Type, string]

// TypeFullName returns the full name of the type.
//
// It contains the full package path if the type is defined in a package.
func TypeFullName(typ reflect.Type) string {
	if typ.PkgPath() == "" {
		// Fast path for known base types.
		name := knownBaseTypeNames[typ.Kind()]
		if name != "" {
			return name
		}
	}
	name, ok := typeFullNameCache.Load(typ)
	if ok {
		return name
	}
	name = typeFullName(typ)
	typeFullNameCache.Store(typ, name)
	return name
}

var knownBaseTypeNames = [...]string{
	reflect.Bool:          reflect.TypeFor[bool]().Name(),
	reflect.Int:           reflect.TypeFor[int]().Name(),
	reflect.Int8:          reflect.TypeFor[int8]().Name(),
	reflect.Int16:         reflect.TypeFor[int16]().Name(),
	reflect.Int32:         reflect.TypeFor[int32]().Name(),
	reflect.Int64:         reflect.TypeFor[int64]().Name(),
	reflect.Uint:          reflect.TypeFor[uint]().Name(),
	reflect.Uint8:         reflect.TypeFor[uint8]().Name(),
	reflect.Uint16:        reflect.TypeFor[uint16]().Name(),
	reflect.Uint32:        reflect.TypeFor[uint32]().Name(),
	reflect.Uint64:        reflect.TypeFor[uint64]().Name(),
	reflect.Uintptr:       reflect.TypeFor[uintptr]().Name(),
	reflect.Float32:       reflect.TypeFor[float32]().Name(),
	reflect.Float64:       reflect.TypeFor[float64]().Name(),
	reflect.Complex64:     reflect.TypeFor[complex64]().Name(),
	reflect.Complex128:    reflect.TypeFor[complex128]().Name(),
	reflect.String:        reflect.TypeFor[string]().Name(),
	reflect.UnsafePointer: reflect.TypeFor[unsafe.Pointer]().Name(),
}

func typeFullName(typ reflect.Type) string {
	pkgPath := typ.PkgPath()
	if pkgPath != "" {
		return pkgPath + "." + typ.Name()
	}
	switch typ.Kind() {
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

// TypeFullNameFor returns the full name of type parameter.
func TypeFullNameFor[T any]() string {
	s := reflect.TypeFor[typeContainer[T]]().String()
	return s[typeContainerPrefixLen : len(s)-typeContainerSuffixLen]
}

type typeContainer[T any] struct{}

var (
	typeContainerSuffixLen = len("]")
	typeContainerPrefixLen = len(reflect.TypeFor[typeContainer[string]]().String()) - len("string") - typeContainerSuffixLen
)
