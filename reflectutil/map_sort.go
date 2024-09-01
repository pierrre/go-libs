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
