//nolint:dupl // Similar but not duplicated.
package reflectutil

import (
	"iter"
	"reflect"

	"github.com/pierrre/go-libs/syncutil"
)

// StructFields represents a list of [reflect.StructField]s of a struct type.
type StructFields struct {
	fs     []reflect.StructField
	byName map[string]reflect.StructField
}

// Len returns the number of [reflect.StructField]s in the [StructFields].
func (fs StructFields) Len() int {
	return len(fs.fs)
}

// Get returns the [reflect.StructField] at the given index.
func (fs StructFields) Get(i int) reflect.StructField {
	return fs.fs[i]
}

// GetByName returns the [reflect.StructField] with the given name.
func (fs StructFields) GetByName(name string) (reflect.StructField, bool) {
	f, ok := fs.byName[name]
	return f, ok
}

// Range iterates over all [reflect.StructField]s in the [StructFields] and calls the given yield function.
func (fs StructFields) Range(yield func(int, reflect.StructField) bool) {
	for i, f := range fs.fs {
		if !yield(i, f) {
			break
		}
	}
}

// All returns an [iter.Seq2] that iterates over all [reflect.StructField]s in the [StructFields].
func (fs StructFields) All() iter.Seq2[int, reflect.StructField] {
	return fs.Range
}

var structFieldsCache syncutil.Map[reflect.Type, StructFields]

// GetStructFields returns a [StructFields] containing all [reflect.StructField]s of the given type.
// If the type is nil or has no fields, it returns an empty [StructFields].
func GetStructFields(typ reflect.Type) StructFields {
	l := typ.NumField()
	if l == 0 {
		return StructFields{}
	}
	fs, ok := structFieldsCache.Load(typ)
	if ok {
		return fs
	}
	fs.fs = make([]reflect.StructField, l)
	fs.byName = make(map[string]reflect.StructField, l)
	for i := range l {
		f := typ.Field(i)
		fs.fs[i] = f
		fs.byName[f.Name] = f
	}
	fs, _ = structFieldsCache.LoadOrStore(typ, fs)
	return fs
}
