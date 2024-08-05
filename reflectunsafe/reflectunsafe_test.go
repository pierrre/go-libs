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
	}).Elem()
	v = v.FieldByName("unexported")
	assert.False(t, v.CanInterface())
	assert.True(t, v.CanAddr())
	v = ConvertValueCanInterface(v)
	assert.True(t, v.CanInterface())
	assert.True(t, v.CanAddr())
}

func TestConvertValueCanInterfaceAlreadyOK(t *testing.T) {
	v := reflect.ValueOf("test")
	assert.True(t, v.CanInterface())
	assert.False(t, v.CanAddr())
	v = ConvertValueCanInterface(v)
	assert.True(t, v.CanInterface())
	assert.False(t, v.CanAddr())
}

func TestConvertValueCanInterfaceNotAddressable(t *testing.T) {
	v := reflect.ValueOf(testStruct{
		unexported: "test",
	})
	v = v.FieldByName("unexported")
	assert.False(t, v.CanInterface())
	assert.False(t, v.CanAddr())
	v = ConvertValueCanInterface(v)
	assert.False(t, v.CanInterface())
	assert.False(t, v.CanAddr())
}

type testStruct struct {
	unexported string
}
