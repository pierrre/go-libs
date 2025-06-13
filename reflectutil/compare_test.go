package reflectutil_test

import (
	"cmp"
	"reflect"
	"runtime"
	"slices"
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/reflectutil"
)

var compareTestCases = []struct {
	name     string
	a, b     any
	expected int
}{
	{
		name:     "DifferentKinds",
		a:        1,
		b:        "a",
		expected: -1,
	},
	{
		name: "DifferentTypes",
		a:    1,
		b: func() any {
			type customInt int
			return customInt(1)
		}(),
		expected: -1,
	},
	{
		name:     "BoolLess",
		a:        false,
		b:        true,
		expected: -1,
	},
	{
		name:     "BoolGreater",
		a:        true,
		b:        false,
		expected: 1,
	},
	{
		name:     "BoolEqualTrue",
		a:        true,
		b:        true,
		expected: 0,
	},
	{
		name:     "BoolEqualFalse",
		a:        false,
		b:        false,
		expected: 0,
	},
	{
		name:     "IntLess",
		a:        1,
		b:        2,
		expected: -1,
	},
	{
		name:     "IntGreater",
		a:        2,
		b:        1,
		expected: 1,
	},
	{
		name:     "IntEqual",
		a:        1,
		b:        1,
		expected: 0,
	},
	{
		name:     "UintLess",
		a:        uint(1),
		b:        uint(2),
		expected: -1,
	},
	{
		name:     "UintGreater",
		a:        uint(2),
		b:        uint(1),
		expected: 1,
	},
	{
		name:     "UintEqual",
		a:        uint(1),
		b:        uint(1),
		expected: 0,
	},
	{
		name:     "FloatLess",
		a:        1.0,
		b:        2.0,
		expected: -1,
	},
	{
		name:     "FloatGreater",
		a:        2.0,
		b:        1.0,
		expected: 1,
	},
	{
		name:     "FloatEqual",
		a:        1.0,
		b:        1.0,
		expected: 0,
	},
	{
		name:     "ComplexRealLess",
		a:        1 + 1i,
		b:        2 + 1i,
		expected: -1,
	},
	{
		name:     "ComplexRealGreater",
		a:        2 + 1i,
		b:        1 + 1i,
		expected: 1,
	},
	{
		name:     "ComplexImagLess",
		a:        1 + 1i,
		b:        1 + 2i,
		expected: -1,
	},
	{
		name:     "ComplexImagGreater",
		a:        1 + 2i,
		b:        1 + 1i,
		expected: 1,
	},
	{
		name:     "ComplexEqual",
		a:        1 + 1i,
		b:        1 + 1i,
		expected: 0,
	},
	{
		name:     "ArrayLess",
		a:        [2]int{1, 1},
		b:        [2]int{1, 2},
		expected: -1,
	},
	{
		name:     "ArrayGreater",
		a:        [2]int{1, 2},
		b:        [2]int{1, 1},
		expected: 1,
	},
	{
		name:     "ArrayEqual",
		a:        [2]int{1, 1},
		b:        [2]int{1, 1},
		expected: 0,
	},
	{
		name:     "ChanLess",
		a:        testChans[0],
		b:        testChans[1],
		expected: -1,
	},
	{
		name:     "ChanGreater",
		a:        testChans[1],
		b:        testChans[0],
		expected: 1,
	},
	{
		name:     "ChanEqual",
		a:        testChans[0],
		b:        testChans[0],
		expected: 0,
	},
	{
		name:     "FuncLess",
		a:        testFuncs[0],
		b:        testFuncs[1],
		expected: -1,
	},
	{
		name:     "FuncGreater",
		a:        testFuncs[1],
		b:        testFuncs[0],
		expected: 1,
	},
	{
		name:     "FuncEqual",
		a:        testFuncs[0],
		b:        testFuncs[0],
		expected: 0,
	},
	{
		name:     "InterfaceNilLess",
		a:        [1]any{nil},
		b:        [1]any{1},
		expected: -1,
	},
	{
		name:     "InterfaceNilGreater",
		a:        [1]any{1},
		b:        [1]any{nil},
		expected: 1,
	},
	{
		name:     "InterfaceNilEqual",
		a:        [1]any{nil},
		b:        [1]any{nil},
		expected: 0,
	},
	{
		name:     "InterfaceTypeLess",
		a:        [1]any{"a"},
		b:        [1]any{1},
		expected: -1,
	},
	{
		name:     "InterfaceTypeGreater",
		a:        [1]any{1},
		b:        [1]any{"a"},
		expected: 1,
	},
	{
		name:     "InterfaceValueLess",
		a:        [1]any{1},
		b:        [1]any{2},
		expected: -1,
	},
	{
		name:     "InterfaceValueGreater",
		a:        [1]any{2},
		b:        [1]any{1},
		expected: 1,
	},
	{
		name:     "InterfaceValueEqual",
		a:        [1]any{1},
		b:        [1]any{1},
		expected: 0,
	},
	{
		name:     "MapLess",
		a:        testMaps[0],
		b:        testMaps[1],
		expected: -1,
	},
	{
		name:     "MapGreater",
		a:        testMaps[1],
		b:        testMaps[0],
		expected: 1,
	},
	{
		name:     "MapEqual",
		a:        testMaps[0],
		b:        testMaps[0],
		expected: 0,
	},
	{
		name:     "PointerLess",
		a:        &testInts[0],
		b:        &testInts[1],
		expected: -1,
	},
	{
		name:     "PointerGreater",
		a:        &testInts[1],
		b:        &testInts[0],
		expected: 1,
	},
	{
		name:     "PointerEqual",
		a:        &testInts[0],
		b:        &testInts[0],
		expected: 0,
	},
	{
		name:     "SliceLess",
		a:        testSlices[0],
		b:        testSlices[1],
		expected: -1,
	},
	{
		name:     "SliceGreater",
		a:        testSlices[1],
		b:        testSlices[0],
		expected: 1,
	},
	{
		name:     "SliceEqual",
		a:        testSlices[0],
		b:        testSlices[0],
		expected: 0,
	},
	{
		name:     "BytesLess",
		a:        []byte("a"),
		b:        []byte("b"),
		expected: -1,
	},
	{
		name:     "BytesGreater",
		a:        []byte("b"),
		b:        []byte("a"),
		expected: 1,
	},
	{
		name:     "BytesEqual",
		a:        []byte("a"),
		b:        []byte("a"),
		expected: 0,
	},
	{
		name:     "StringLess",
		a:        "a",
		b:        "b",
		expected: -1,
	},
	{
		name:     "StringGreater",
		a:        "b",
		b:        "a",
		expected: 1,
	},
	{
		name:     "StringEqual",
		a:        "a",
		b:        "a",
		expected: 0,
	},
	{
		name:     "StructLess",
		a:        struct{ A, B int }{1, 1},
		b:        struct{ A, B int }{1, 2},
		expected: -1,
	},
	{
		name:     "StructGreater",
		a:        struct{ A, B int }{1, 2},
		b:        struct{ A, B int }{1, 1},
		expected: 1,
	},
	{
		name:     "StructEqual",
		a:        struct{ A, B int }{1, 1},
		b:        struct{ A, B int }{1, 1},
		expected: 0,
	},
}

func TestCompare(t *testing.T) {
	for _, tc := range compareTestCases {
		ra := reflect.ValueOf(tc.a)
		rb := reflect.ValueOf(tc.b)
		t.Run(tc.name, func(t *testing.T) {
			c := Compare(ra, rb)
			assert.Equal(t, c, tc.expected)
			allocs := testing.AllocsPerRun(100, func() {
				Compare(ra, rb)
			})
			assert.Equal(t, allocs, 0)
		})
	}
}

func BenchmarkCompare(b *testing.B) {
	for _, tc := range compareTestCases {
		ra := reflect.ValueOf(tc.a)
		rb := reflect.ValueOf(tc.b)
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				Compare(ra, rb)
			}
		})
	}
}

func BenchmarkComparison(b *testing.B) {
	for _, tc := range compareTestCases {
		ra := reflect.ValueOf(tc.a)
		rb := reflect.ValueOf(tc.b)
		if ra.Type() != rb.Type() {
			continue
		}
		f := GetCompareFunc(ra.Type())
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				f(ra, rb)
			}
		})
	}
}

var (
	testInts   = [2]int{}
	testChans  = makeTestChans()
	testFuncs  = makeTestFuncs()
	testMaps   = makeTestMaps()
	testSlices = makeTestSlices()
	testPin    runtime.Pinner
)

func makeTestChans() []chan int {
	cs := []chan int{make(chan int), make(chan int)}
	for i := range cs {
		testPin.Pin(reflect.ValueOf(cs[i]).UnsafePointer())
	}
	slices.SortFunc(cs, func(a, b chan int) int {
		return cmp.Compare(reflect.ValueOf(a).Pointer(), reflect.ValueOf(b).Pointer())
	})
	return cs
}

func makeTestFuncs() []func() {
	fs := []func(){func() {}, func() {}}
	for i := range fs {
		testPin.Pin(reflect.ValueOf(fs[i]).UnsafePointer())
	}
	slices.SortFunc(fs, func(a, b func()) int {
		return cmp.Compare(reflect.ValueOf(a).Pointer(), reflect.ValueOf(b).Pointer())
	})
	return fs
}

func makeTestMaps() []map[int]int {
	ms := []map[int]int{{}, {}}
	for i := range ms {
		testPin.Pin(reflect.ValueOf(ms[i]).UnsafePointer())
	}
	slices.SortFunc(ms, func(a, b map[int]int) int {
		return cmp.Compare(reflect.ValueOf(a).Pointer(), reflect.ValueOf(b).Pointer())
	})
	return ms
}

func makeTestSlices() [][]int {
	ss := [][]int{{1}, {2}}
	for i := range ss {
		testPin.Pin(reflect.ValueOf(ss[i]).UnsafePointer())
	}
	slices.SortFunc(ss, func(a, b []int) int {
		return cmp.Compare(reflect.ValueOf(a).Pointer(), reflect.ValueOf(b).Pointer())
	})
	return ss
}
