package reflectunsafe_test

import (
	"reflect"
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/reflectunsafe"
)

func TestConvertValueCanInterfaceOK(t *testing.T) {
	v := reflect.ValueOf(&testStruct{
		unexported: "test",
	}).Elem().FieldByName("unexported")
	assert.False(t, v.CanInterface())
	v = ConvertValueCanInterface(v)
	assert.True(t, v.CanInterface())
	vi := v.Interface()
	s, _ := assert.Type[string](t, vi)
	assert.Equal(t, s, "test")
}

func TestConvertValueCanInterfacePointer(t *testing.T) {
	s1 := "test"
	p1 := &s1
	v := reflect.ValueOf(&testStruct{
		pointer: p1,
	}).Elem().FieldByName("pointer")
	assert.False(t, v.CanInterface())
	v = ConvertValueCanInterface(v)
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
	v := reflect.ValueOf(&testStruct{
		pointerPointer: pp1,
	}).Elem().FieldByName("pointerPointer")
	assert.False(t, v.CanInterface())
	v = ConvertValueCanInterface(v)
	assert.True(t, v.CanInterface())
	vi := v.Interface()
	pp2, _ := assert.Type[**string](t, vi)
	assert.NotZero(t, pp2)
	p2 := *pp2
	assert.NotZero(t, p2)
	s2 := *p2
	assert.Equal(t, s2, "test")
}

func TestConvertValueCanInterfaceAlreadyOK(t *testing.T) {
	v := reflect.ValueOf("test")
	assert.True(t, v.CanInterface())
	v = ConvertValueCanInterface(v)
	assert.True(t, v.CanInterface())
	vi := v.Interface()
	s, _ := assert.Type[string](t, vi)
	assert.Equal(t, s, "test")
}

func TestConvertValueCanInterfaceNotAddressable(t *testing.T) {
	v := reflect.ValueOf(testStruct{
		unexported: "test",
	}).FieldByName("unexported")
	assert.False(t, v.CanInterface())
	v = ConvertValueCanInterface(v)
	assert.False(t, v.CanInterface())
}

type testStruct struct {
	unexported     string
	pointer        *string
	pointerPointer **string
}
