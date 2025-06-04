//nolint:dupl // Similar but not duplicated.
package reflectutil

import (
	"iter"
	"reflect"

	"github.com/pierrre/go-libs/syncutil"
)

// Methods represents a list of [reflect.Method]s of a type.
type Methods struct {
	ms     []reflect.Method
	byName map[string]reflect.Method
}

// Len returns the number of [reflect.Method] in the [Methods].
func (ms Methods) Len() int {
	return len(ms.ms)
}

// Get returns the [reflect.Method] at the given index.
func (ms Methods) Get(i int) reflect.Method {
	return ms.ms[i]
}

// GetByName returns the [reflect.Method] with the given name.
func (ms Methods) GetByName(name string) (reflect.Method, bool) {
	m, ok := ms.byName[name]
	return m, ok
}

// Range iterates over all [reflect.Method]s in the [Methods] and calls the given yield function.
func (ms Methods) Range(yield func(int, reflect.Method) bool) {
	for i, m := range ms.ms {
		if !yield(i, m) {
			break
		}
	}
}

// All returns an [iter.Seq2] that iterates over all [reflect.Method]s in the [Methods].
func (ms Methods) All() iter.Seq2[int, reflect.Method] {
	return ms.Range
}

var methodsCache syncutil.Map[reflect.Type, Methods]

// GetMethods returns a [Methods] containing all [reflect.Method]s of the given type.
// If the type is nil or has no methods, it returns an empty [Methods].
func GetMethods(typ reflect.Type) Methods {
	l := typ.NumMethod()
	if l == 0 {
		return Methods{}
	}
	ms, ok := methodsCache.Load(typ)
	if ok {
		return ms
	}
	ms.ms = make([]reflect.Method, l)
	ms.byName = make(map[string]reflect.Method, l)
	for i := range l {
		m := typ.Method(i)
		ms.ms[i] = m
		ms.byName[m.Name] = m
	}
	ms, _ = methodsCache.LoadOrStore(typ, ms)
	return ms
}
