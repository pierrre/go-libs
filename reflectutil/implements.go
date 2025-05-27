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
	itf               reflect.Type
	numMethod         int
	numMethodExported int
	m                 syncutil.Map[reflect.Type, bool]
}

// NewImplementsCache creates a new [ImplementsCache] for the given interface type.
func NewImplementsCache(itf reflect.Type) *ImplementsCache {
	if itf.Kind() != reflect.Interface {
		panic(fmt.Errorf("not interface: %s", itf.Kind()))
	}
	c := &ImplementsCache{
		itf: itf,
	}
	c.numMethod = itf.NumMethod()
	for i := range c.numMethod {
		m := itf.Method(i)
		if m.IsExported() {
			c.numMethodExported++
		}
	}
	return c
}

// NewImplementsCacheFor creates a new [ImplementsCache] for the given type parameter.
func NewImplementsCacheFor[T any]() *ImplementsCache {
	return NewImplementsCache(reflect.TypeFor[T]())
}

// ImplementedBy checks if the interface is impleented by the given type.
func (c *ImplementsCache) ImplementedBy(typ reflect.Type) bool {
	if typ == nil {
		return false
	}
	if c.numMethod == 0 {
		return true
	}
	if typ == c.itf {
		return true
	}
	numMethod := c.numMethod
	if typ.Kind() != reflect.Interface {
		numMethod = c.numMethodExported
	}
	if typ.NumMethod() < numMethod {
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
