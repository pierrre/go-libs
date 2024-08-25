package reflectutil_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/pierrre/assert"
	"github.com/pierrre/assert/assertauto"
	. "github.com/pierrre/go-libs/reflectutil"
)

var benchRes any

type typeFullNameVariantTestCase struct {
	name          string
	newFunc       func(tc typeFullNameTestCase) func() string
	benchParallel bool
}

var typeFullNameVariantTestCases = []typeFullNameVariantTestCase{
	{
		name: "Normal",
		newFunc: func(tc typeFullNameTestCase) func() string {
			return func() string {
				return TypeFullName(tc.typ)
			}
		},
		benchParallel: true,
	},
	{
		name: "Internal",
		newFunc: func(tc typeFullNameTestCase) func() string {
			return func() string {
				return TypeFullNameInternal(tc.typ)
			}
		},
	},
	{
		name: "For",
		newFunc: func(tc typeFullNameTestCase) func() string {
			return tc.forFunc
		},
	},
}

func runTypeFullNameVariantTestCases[TB interface {
	testing.TB
	Run(name string, f func(TB)) bool
}](tb TB, f func(tb TB, variantTC typeFullNameVariantTestCase, tc typeFullNameTestCase)) {
	for _, variantTC := range typeFullNameVariantTestCases {
		tb.Run(variantTC.name, func(tb TB) {
			runTypeFullNameTestCases(tb, func(tb TB, tc typeFullNameTestCase) {
				f(tb, variantTC, tc)
			})
		})
	}
}

type typeFullNameTestCase struct {
	typ     reflect.Type
	forFunc func() string
}

func newTypeFullNameTestCase[T any]() typeFullNameTestCase {
	return typeFullNameTestCase{
		typ:     reflect.TypeFor[T](),
		forFunc: TypeFullNameFor[T],
	}
}

var typeFullNameTestCases = []typeFullNameTestCase{
	newTypeFullNameTestCase[string](),
	newTypeFullNameTestCase[**********string](),
	newTypeFullNameTestCase[<-chan map[string][][2]*string](),
	newTypeFullNameTestCase[testType](),
	newTypeFullNameTestCase[*testType](),
	newTypeFullNameTestCase[<-chan map[string][][2]*testType](),
	newTypeFullNameTestCase[testPointer](),
	newTypeFullNameTestCase[*testPointer](),
	newTypeFullNameTestCase[<-chan map[string][][2]*testPointer](),
	newTypeFullNameTestCase[testContainer[testType]](),
	newTypeFullNameTestCase[*testContainer[testType]](),
	newTypeFullNameTestCase[<-chan map[string][][2]*testContainer[chan map[string][][2]*testType]](),
}

func runTypeFullNameTestCases[TB interface {
	testing.TB
	Run(name string, f func(TB)) bool
}](tb TB, f func(tb TB, tc typeFullNameTestCase)) {
	for _, tc := range typeFullNameTestCases {
		f(tb, tc)
	}
}

type testType struct{}

type testContainer[T any] struct{}

type testPointer *testType

func getTypeFullNameTestName(typ reflect.Type) string {
	return strings.ReplaceAll(TypeFullName(typ), "/", "_")
}

func TestTypeFullName(t *testing.T) {
	runTypeFullNameVariantTestCases(t, func(t *testing.T, variantTC typeFullNameVariantTestCase, tc typeFullNameTestCase) { //nolint:thelper // This is not a test helper.
		f := variantTC.newFunc(tc)
		assertauto.Equal(t, f(), assertauto.Name("name"))
		assertauto.AllocsPerRun(t, 100, func() {
			_ = f()
		}, assertauto.Name("allocs"))
	})
}

func BenchmarkTypeFullName(b *testing.B) {
	runTypeFullNameVariantTestCases(b, func(b *testing.B, variantTC typeFullNameVariantTestCase, tc typeFullNameTestCase) { //nolint:thelper // This is not a test helper.
		f := variantTC.newFunc(tc)
		b.Run(getTypeFullNameTestName(tc.typ), func(b *testing.B) {
			benchSeq := func(b *testing.B) { //nolint:thelper // This is not a test helper.
				var res string
				for range b.N {
					res = f()
				}
				benchRes = res
			}
			if variantTC.benchParallel {
				b.Run("Sequential", benchSeq)
				b.Run("Parallel", func(b *testing.B) {
					b.RunParallel(func(pb *testing.PB) {
						var res string
						for pb.Next() {
							res = f()
						}
						benchRes = res
					})
				})
			} else {
				benchSeq(b)
			}
		})
	})
}

func TestValueInterfaceFor(t *testing.T) {
	s1 := "test"
	v := reflect.ValueOf(s1)
	s2 := ValueInterfaceFor[string](v)
	assert.Equal(t, s1, s2)
}

func TestConvertValueCanInterfaceAlreadyOK(t *testing.T) {
	v := reflect.ValueOf("test")
	assert.True(t, v.CanInterface())
	v, ok := ConvertValueCanInterface(v)
	assert.True(t, ok)
	assert.True(t, v.CanInterface())
	vi := v.Interface()
	s, _ := assert.Type[string](t, vi)
	assert.Equal(t, s, "test")
}

func TestConvertValueCanInterfacePointer(t *testing.T) {
	s1 := "test"
	p1 := &s1
	v := reflect.ValueOf(testStruct{
		pointer: p1,
	}).FieldByName("pointer")
	assert.False(t, v.CanInterface())
	v, ok := ConvertValueCanInterface(v)
	assert.True(t, ok)
	assert.True(t, v.CanInterface())
	vi := v.Interface()
	p2, _ := assert.Type[*string](t, vi)
	assert.NotZero(t, p2)
	s2 := *p2
	assert.Equal(t, s2, "test")
}

func TestConvertValueCanInterfacePointerPointer(t *testing.T) {
	s1 := "test"
	p1 := &s1
	pp1 := &p1
	v := reflect.ValueOf(testStruct{
		pointerPointer: pp1,
	}).FieldByName("pointerPointer")
	assert.False(t, v.CanInterface())
	v, ok := ConvertValueCanInterface(v)
	assert.True(t, ok)
	assert.True(t, v.CanInterface())
	vi := v.Interface()
	pp2, _ := assert.Type[**string](t, vi)
	assert.NotZero(t, pp2)
	p2 := *pp2
	assert.NotZero(t, p2)
	s2 := *p2
	assert.Equal(t, s2, "test")
}

func TestConvertValueCanInterfaceAddressable(t *testing.T) {
	v := reflect.ValueOf(&testStruct{
		unexported: "test",
	}).Elem().FieldByName("unexported")
	assert.False(t, v.CanInterface())
	v, ok := ConvertValueCanInterface(v)
	assert.True(t, ok)
	assert.True(t, v.CanInterface())
	vi := v.Interface()
	s, _ := assert.Type[string](t, vi)
	assert.Equal(t, s, "test")
}

func TestConvertValueCanInterfaceFail(t *testing.T) {
	v := reflect.ValueOf(testStruct{
		unexported: "test",
	}).FieldByName("unexported")
	assert.False(t, v.CanInterface())
	v, ok := ConvertValueCanInterface(v)
	assert.False(t, ok)
	assert.False(t, v.CanInterface())
}

func TestTryValueInterfaceOK(t *testing.T) {
	s1 := "test"
	v := reflect.ValueOf(s1)
	vi, ok := TryValueInterface(v)
	assert.True(t, ok)
	s2, _ := assert.Type[string](t, vi)
	assert.Equal(t, s1, s2)
}

func TestTryValueInterfaceFail(t *testing.T) {
	s1 := "test"
	v := reflect.ValueOf(testStruct{
		unexported: s1,
	}).FieldByName("unexported")
	vi, ok := TryValueInterface(v)
	assert.False(t, ok)
	assert.Zero(t, vi)
}

type testStruct struct {
	unexported     string
	pointer        *string
	pointerPointer **string
}

func TestTryValueInterfaceForOK(t *testing.T) {
	s1 := "test"
	v := reflect.ValueOf(s1)
	vi, ok := TryValueInterfaceFor[string](v)
	assert.True(t, ok)
	s2, _ := assert.Type[string](t, vi)
	assert.Equal(t, s1, s2)
}

func TestTryValueInterfaceForFail(t *testing.T) {
	s1 := "test"
	v := reflect.ValueOf(testStruct{
		unexported: s1,
	}).FieldByName("unexported")
	s2, ok := TryValueInterfaceFor[string](v)
	assert.False(t, ok)
	assert.Zero(t, s2)
}
