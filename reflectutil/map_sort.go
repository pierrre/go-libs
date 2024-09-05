package reflectutil

import (
	"reflect"
	"slices"
)

// GetSortedMap returns a sorted [MapEntries] of the given map.
func GetSortedMap(m reflect.Value) MapEntries {
	es := GetMapEntries(m)
	cmpFunc := GetCompareFunc(m.Type().Key())
	slices.SortFunc(es, func(a, b MapEntry) int {
		return cmpFunc(a.Key, b.Key)
	})
	return es
}

// GetSortedMapKeys returns a sorted [MapKeys] of the given map.
func GetSortedMapKeys(m reflect.Value) MapKeys {
	ks := GetMapKeys(m)
	cmpFunc := GetCompareFunc(m.Type().Key())
	slices.SortFunc(ks, cmpFunc)
	return ks
}
