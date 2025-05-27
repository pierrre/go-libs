package reflectutil

import (
	"fmt"
	"reflect"

	"github.com/pierrre/go-libs/syncutil"
)

// ImplementsCache is a cache for checking if an interface is implemented by types.
//
// It should be created with [NewImplementsCache].
type ImplementsCache struct {
	itf         reflect.Type
	numExported int
	m           syncutil.Map[reflect.Type, bool]
}

// NewImplementsCache creates a new [ImplementsCache] for the given interface type.
func NewImplementsCache(itf reflect.Type) *ImplementsCache {
	if itf.Kind() != reflect.Interface {
		panic(fmt.Errorf("not interface: %s", itf.Kind()))
	}
	numExported := 0
	for i := range itf.NumMethod() {
		m := itf.Method(i)
		if m.IsExported() {
			numExported++
		}
	}
	return &ImplementsCache{
		itf:         itf,
		numExported: numExported,
	}
}

// NewImplementsCacheFor creates a new [ImplementsCache] for the given type parameter.
func NewImplementsCacheFor[T any]() *ImplementsCache {
	return NewImplementsCache(reflect.TypeFor[T]())
}

// ImplementedBy checks if the interface is implemented by the given type.
func (c *ImplementsCache) ImplementedBy(typ reflect.Type) bool {
	if typ == nil {
		return false
	}
	if typ == c.itf {
		return true
	}
	if typ.NumMethod() < c.numExported {
		// It's faster than loading from the map.
		return false
	}
	implements, ok := c.m.Load(typ)
	if ok {
		return implements
	}
	implements = typ.Implements(c.itf)
	c.m.Store(typ, implements)
	return implements
}
