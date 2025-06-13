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
		name:     "ChanNilLess",
		a:        chan int(nil),
		b:        make(chan int),
		expected: -1,
	},
	{
		name:     "ChanNilGreater",
		a:        make(chan int),
		b:        chan int(nil),
		expected: 1,
	},
	{
		name:     "ChanNilEqual",
		a:        chan int(nil),
		b:        chan int(nil),
		expected: 0,
	},
	{
		name:     "ChanPointerLess",
		a:        chans[0],
		b:        chans[1],
		expected: -1,
	},
	{
		name:     "ChanPointerGreater",
		a:        chans[1],
		b:        chans[0],
		expected: 1,
	},
	{
		name:     "ChanPointerEqual",
		a:        chans[0],
		b:        chans[0],
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
		name:     "PointerLess",
		a:        &ints[0],
		b:        &ints[1],
		expected: -1,
	},
	{
		name:     "PointerGreater",
		a:        &ints[1],
		b:        &ints[0],
		expected: 1,
	},
	{
		name:     "PointerEqual",
		a:        &ints[0],
		b:        &ints[0],
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

func TestComparePanicUnsupportedType(t *testing.T) {
	assert.Panics(t, func() {
		Compare(reflect.ValueOf(func() {}), reflect.ValueOf(func() {}))
	})
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
	ints  = [2]int{}
	chans = makeChans()
	pin   runtime.Pinner
)

func makeChans() []chan int {
	cs := []chan int{make(chan int), make(chan int)}
	for i := range cs {
		pin.Pin(reflect.ValueOf(cs[i]).UnsafePointer())
	}
	slices.SortFunc(cs, func(a, b chan int) int {
		return cmp.Compare(reflect.ValueOf(a).Pointer(), reflect.ValueOf(b).Pointer())
	})
	return cs
}
