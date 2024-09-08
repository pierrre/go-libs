package reflectutil

import (
	"reflect"
	"slices"

	"github.com/pierrre/go-libs/syncutil"
)

// MapEntry represents a key-value pair in a map.
type MapEntry struct {
	Key   reflect.Value
	Value reflect.Value
}

// MapEntries is a slice of [MapEntry].
type MapEntries []MapEntry

// GetMapEntries returns the entries of a map.
func GetMapEntries(m reflect.Value) MapEntries {
	if m.Len() == 0 {
		return nil
	}
	if m.CanInterface() {
		return getMapEntriesExported(m)
	}
	return getMapEntriesUnexported(m)
}

func getMapEntriesExported(m reflect.Value) MapEntries {
	typ := m.Type()
	keyTyp := typ.Key()
	elemTyp := typ.Elem()
	es := getMapEntriesFromPool(typ)
	es = es[:0]
	es = slices.Grow(es, m.Len())
	es = es[:m.Len()]
	iter := m.MapRange()
	for i := 0; iter.Next(); i++ {
		e := es[i]
		if !e.Key.IsValid() {
			e.Key = reflect.New(keyTyp).Elem()
		}
		e.Key.SetIterKey(iter)
		if !e.Value.IsValid() {
			e.Value = reflect.New(elemTyp).Elem()
		}
		e.Value.SetIterValue(iter)
		es[i] = e
	}
	return es
}

func getMapEntriesUnexported(m reflect.Value) MapEntries {
	es := make(MapEntries, 0, m.Len())
	iter := m.MapRange()
	for iter.Next() {
		es = append(es, MapEntry{
			Key:   iter.Key(),
			Value: iter.Value(),
		})
	}
	return es
}

// Release releases the [MapEntries].
//
// It helps to reduce memory allocations.
func (es MapEntries) Release() {
	if len(es) == 0 {
		return
	}
	e := es[0]
	if !e.Key.CanInterface() || !e.Value.CanInterface() {
		return
	}
	for i, e := range es {
		e.Key.SetZero()
		e.Value.SetZero()
		es[i] = e
	}
	typ := reflect.MapOf(e.Key.Type(), e.Value.Type())
	putMapEntriesToPool(typ, es)
}

var mapEntriesPools = syncutil.Map[reflect.Type, *syncutil.ValuePool[MapEntries]]{}

func getMapEntriesPool(typ reflect.Type) *syncutil.ValuePool[MapEntries] {
	pool, ok := mapEntriesPools.Load(typ)
	if !ok {
		pool = &syncutil.ValuePool[MapEntries]{}
		mapEntriesPools.Store(typ, pool)
	}
	return pool
}

func getMapEntriesFromPool(typ reflect.Type) MapEntries {
	pool := getMapEntriesPool(typ)
	return pool.Get()
}

func putMapEntriesToPool(typ reflect.Type, entries MapEntries) {
	pool := getMapEntriesPool(typ)
	pool.Put(entries)
}
