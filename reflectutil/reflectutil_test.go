package reflectutil_test

import (
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/pierrre/assert"
	"github.com/pierrre/assert/assertauto"
	. "github.com/pierrre/go-libs/reflectutil"
)

var benchRes any

var types = []reflect.Type{
	reflect.TypeFor[string](),
	reflect.TypeFor[int](),
	reflect.TypeFor[*string](),
	reflect.TypeFor[any](),
	reflect.TypeFor[reflect.Type](),
	reflect.TypeFor[*reflect.Type](),
	reflect.TypeFor[reflect.Value](),
	reflect.TypeFor[*reflect.Value](),
	reflect.TypeFor[atomic.Value](),
	reflect.TypeFor[*atomic.Value](),
}

func TestTypeFullName(t *testing.T) {
	for _, typ := range types {
		s := TypeFullName(typ)
		assertauto.Equal(t, s, assertauto.AssertOptions(assert.MessageWrap(s)))
	}
}

func TestTypeFullNameAllocs(t *testing.T) {
	for _, typ := range types {
		assert.AllocsPerRun(t, 100, func() {
			_ = TypeFullName(typ)
		}, 0, assert.MessageWrap(TypeFullName(typ)))
	}
}

func BenchmarkTypeFullName(b *testing.B) {
	for _, typ := range types {
		b.Run(TypeFullName(typ), func(b *testing.B) {
			var res string
			for range b.N {
				res = TypeFullName(typ)
			}
			benchRes = res
		})
	}
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
