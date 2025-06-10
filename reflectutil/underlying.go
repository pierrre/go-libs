package reflectutil

import (
	"reflect"
	"unsafe" //nolint:depguard // Required for unsafe.Pointer.

	"github.com/pierrre/go-libs/syncutil"
)

// GetUnderlyingType returns the underlying type of the given type.
// E.g. for the type defined as `type MyType string`, it returns `string`.
// It returns the type itself if it's its own underlying type, or if the kind is invalid/interface/struct.
func GetUnderlyingType(typ reflect.Type) reflect.Type {
	kind := typ.Kind()
	uTyp := knownKindUnderlyingTypes[kind]
	if uTyp != nil {
		return uTyp
	}
	switch kind { //nolint:exhaustive // We only need these kinds.
	case reflect.Invalid, reflect.Interface, reflect.Struct:
		return typ
	}
	uTyp, ok := underlyingTypeCache.Load(typ)
	if ok {
		return uTyp
	}
	uTyp = computeUnderlyingType(typ)
	underlyingTypeCache.Store(typ, uTyp)
	return uTyp
}

// GetBaseType is an alias for [GetUnderlyingType].
//
// Deprecated: use [GetUnderlyingType] instead.
func GetBaseType(typ reflect.Type) reflect.Type {
	return GetUnderlyingType(typ)
}

var knownKindUnderlyingTypes = [...]reflect.Type{
	reflect.Bool:          reflect.TypeFor[bool](),
	reflect.Int:           reflect.TypeFor[int](),
	reflect.Int8:          reflect.TypeFor[int8](),
	reflect.Int16:         reflect.TypeFor[int16](),
	reflect.Int32:         reflect.TypeFor[int32](),
	reflect.Int64:         reflect.TypeFor[int64](),
	reflect.Uint:          reflect.TypeFor[uint](),
	reflect.Uint8:         reflect.TypeFor[uint8](),
	reflect.Uint16:        reflect.TypeFor[uint16](),
	reflect.Uint32:        reflect.TypeFor[uint32](),
	reflect.Uint64:        reflect.TypeFor[uint64](),
	reflect.Uintptr:       reflect.TypeFor[uintptr](),
	reflect.Float32:       reflect.TypeFor[float32](),
	reflect.Float64:       reflect.TypeFor[float64](),
	reflect.Complex64:     reflect.TypeFor[complex64](),
	reflect.Complex128:    reflect.TypeFor[complex128](),
	reflect.String:        reflect.TypeFor[string](),
	reflect.UnsafePointer: reflect.TypeFor[unsafe.Pointer](),
}

var underlyingTypeCache syncutil.Map[reflect.Type, reflect.Type]

func computeUnderlyingType(typ reflect.Type) reflect.Type {
	uTyp := typ
	switch typ.Kind() { //nolint:exhaustive // We only need these kinds.
	case reflect.Array:
		uTyp = reflect.ArrayOf(typ.Len(), typ.Elem())
	case reflect.Chan:
		uTyp = reflect.ChanOf(typ.ChanDir(), typ.Elem())
	case reflect.Func:
		in := make([]reflect.Type, typ.NumIn())
		for i := range in {
			in[i] = typ.In(i)
		}
		out := make([]reflect.Type, typ.NumOut())
		for i := range out {
			out[i] = typ.Out(i)
		}
		uTyp = reflect.FuncOf(in, out, typ.IsVariadic())
	case reflect.Map:
		uTyp = reflect.MapOf(typ.Key(), typ.Elem())
	case reflect.Pointer:
		uTyp = reflect.PointerTo(typ.Elem())
	case reflect.Slice:
		uTyp = reflect.SliceOf(typ.Elem())
	}
	return uTyp
}
