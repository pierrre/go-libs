package reflectutil

import (
	"reflect"
	"slices"

	"github.com/pierrre/go-libs/syncutil"
)

// MapKeys represents the keys of a map.
type MapKeys []reflect.Value

// GetMapKeys returns the keys of a map.
func GetMapKeys(m reflect.Value) MapKeys {
	if m.Len() == 0 {
		return nil
	}
	if m.CanInterface() {
		return getMapKeysExported(m)
	}
	return getMapKeysUnexported(m)
}

func getMapKeysExported(m reflect.Value) MapKeys {
	typ := m.Type()
	keyTyp := typ.Key()
	ks := getMapKeysFromPool(keyTyp)
	ks = ks[:0]
	ks = slices.Grow(ks, m.Len())
	ks = ks[:m.Len()]
	iter := m.MapRange()
	for i := 0; iter.Next(); i++ {
		k := ks[i]
		if !k.IsValid() {
			k = reflect.New(keyTyp).Elem()
		}
		k.SetIterKey(iter)
		ks[i] = k
	}
	return ks
}

func getMapKeysUnexported(m reflect.Value) MapKeys {
	ks := make(MapKeys, 0, m.Len())
	iter := m.MapRange()
	for iter.Next() {
		ks = append(ks, iter.Key())
	}
	return ks
}

// Release releases the [MapKeys].
//
// It helps to reduce memory allocations.
func (ks MapKeys) Release() {
	if len(ks) == 0 {
		return
	}
	k := ks[0]
	if !k.CanInterface() {
		return
	}
	for i, k := range ks {
		k.SetZero()
		ks[i] = k
	}
	putMapKeysToPool(k.Type(), ks)
}

var mapKeysPools = syncutil.MapFor[reflect.Type, *syncutil.ValuePool[MapKeys]]{}

func getMapKeysPool(typ reflect.Type) *syncutil.ValuePool[MapKeys] {
	pool, ok := mapKeysPools.Load(typ)
	if !ok {
		pool = &syncutil.ValuePool[MapKeys]{}
		mapKeysPools.Store(typ, pool)
	}
	return pool
}

func getMapKeysFromPool(typ reflect.Type) MapKeys {
	pool := getMapKeysPool(typ)
	return pool.Get()
}

func putMapKeysToPool(typ reflect.Type, es MapKeys) {
	pool := getMapKeysPool(typ)
	pool.Put(es)
}
