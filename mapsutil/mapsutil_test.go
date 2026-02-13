package mapsutil_test

import (
	"fmt"
	"maps"
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/mapsutil"
)

func ExampleSorted() {
	m := map[string]int{
		"c": 3,
		"a": 1,
		"b": 2,
	}
	for k, v := range Sorted(m) {
		fmt.Println(k, v)
	}
	// Output:
	// a 1
	// b 2
	// c 3
}

var testMap = func() map[string]int {
	m := make(map[string]int)
	for i := range 16 {
		m[string(rune('a'+i))] = i + 1
	}
	return m
}()

func TestSorted(t *testing.T) {
	i := 0
	for k, v := range Sorted(testMap) {
		assert.Equal(t, k, string(rune('a'+i)))
		assert.Equal(t, v, i+1)
		i++
	}
	assert.Equal(t, i, len(testMap))
}

func TestSortedDeleteKey(t *testing.T) {
	i := 0
	m := maps.Clone(testMap)
	for k, v := range Sorted(m) {
		if k == "a" {
			delete(m, "c")
		}
		assert.NotEqual(t, k, "c")
		assert.NotEqual(t, v, 3)
		i++
	}
	assert.Equal(t, i, len(testMap)-1)
}

func TestSortedInterrupt(t *testing.T) {
	i := 0
	for range Sorted(testMap) {
		if i == 7 {
			break
		}
		i++
	}
	assert.Equal(t, i, 7)
}
