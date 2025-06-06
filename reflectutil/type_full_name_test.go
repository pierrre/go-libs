package reflectutil_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/pierrre/assert"
	"github.com/pierrre/assert/assertauto"
	. "github.com/pierrre/go-libs/reflectutil"
)

type typeFullNameVariantTestCase struct {
	name    string
	newFunc func(tc typeFullNameTestCase) func() string
}

var typeFullNameVariantTestCases = []typeFullNameVariantTestCase{
	{
		name: "Normal",
		newFunc: func(tc typeFullNameTestCase) func() string {
			return func() string {
				return TypeFullName(tc.typ)
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

func runTypeFullNameTestCases[TB testing.TB](tb TB, f func(tb TB, tc typeFullNameTestCase)) {
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
		assertauto.Equal(t, f())
		assert.AllocsPerRun(t, 100, func() {
			_ = f()
		}, 0)
	})
}

func BenchmarkTypeFullName(b *testing.B) {
	runTypeFullNameVariantTestCases(b, func(b *testing.B, variantTC typeFullNameVariantTestCase, tc typeFullNameTestCase) { //nolint:thelper // This is not a test helper.
		f := variantTC.newFunc(tc)
		b.Run(getTypeFullNameTestName(tc.typ), func(b *testing.B) {
			for b.Loop() {
				f()
			}
		})
	})
}
