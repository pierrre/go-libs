package mapsutil_test

import (
	"fmt"
	"maps"
	"strings"
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/mapsutil"
)

func ExampleSortedByKey() {
	m := map[string]int{
		"c": 3,
		"a": 1,
		"b": 2,
	}
	for k, v := range SortedByKey(m) {
		fmt.Println(k, v)
	}
	// Output:
	// a 1
	// b 2
	// c 3
}

func ExampleSortedByKeyFunc() {
	m := map[string]int{
		"c": 3,
		"a": 1,
		"b": 2,
	}
	for k, v := range SortedByKeyFunc(m, func(a, b string) int {
		return -strings.Compare(a, b)
	}) {
		fmt.Println(k, v)
	}
	// Output:
	// c 3
	// b 2
	// a 1
}

var testMap = func() map[string]int {
	m := make(map[string]int)
	for i := range 16 {
		m[string(rune('a'+i))] = i + 1
	}
	return m
}()

func TestSortedByKey(t *testing.T) {
	i := 0
	for k, v := range SortedByKey(testMap) {
		assert.Equal(t, k, string(rune('a'+i)))
		assert.Equal(t, v, i+1)
		i++
	}
	assert.Equal(t, i, len(testMap))
}

func TestSortedByKeyDeleteKey(t *testing.T) {
	i := 0
	m := maps.Clone(testMap)
	for k, v := range SortedByKey(m) {
		if k == "a" {
			delete(m, "c")
		}
		assert.NotEqual(t, k, "c")
		assert.NotEqual(t, v, 3)
		i++
	}
	assert.Equal(t, i, len(testMap)-1)
}

func TestSortedByKeyInterrupt(t *testing.T) {
	i := 0
	for range SortedByKey(testMap) {
		if i == 7 {
			break
		}
		i++
	}
	assert.Equal(t, i, 7)
}

func TestSortedByKeyFunc(t *testing.T) {
	i := 0
	for k, v := range SortedByKeyFunc(testMap, func(a, b string) int {
		return -strings.Compare(a, b)
	}) {
		assert.Equal(t, k, string(rune('a'+len(testMap)-1-i)))
		assert.Equal(t, v, len(testMap)-i)
		i++
	}
	assert.Equal(t, i, len(testMap))
}
