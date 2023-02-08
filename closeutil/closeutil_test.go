package closeutil

import (
	"errors"
	"testing"

	"github.com/pierrre/assert"
	"github.com/pierrre/go-libs/internal/golibstest"
)

func init() {
	golibstest.Configure()
}

func TestFNilNil(t *testing.T) {
	var c1 F
	c2 := c1.Nil()
	c2()
}

func TestFNilNotNil(t *testing.T) {
	var c1 F
	c2 := c1.Nil()
	called := false
	c1 = func() {
		called = true
	}
	c2()
	assert.True(t, called)
}

func TestErrConvertWithError(t *testing.T) {
	ceCalled := false
	ce := Err(func() error {
		ceCalled = true
		return errors.New("error")
	})
	coe := ce.Convert()
	oeCalled := false
	oe := func(err error) {
		oeCalled = true
		assert.Error(t, err)
	}
	coe(oe)
	assert.True(t, ceCalled)
	assert.True(t, oeCalled)
}

func TestErrConvertNoError(t *testing.T) {
	ceCalled := false
	ce := Err(func() error {
		ceCalled = true
		return nil
	})
	coe := ce.Convert()
	oe := func(err error) {
		t.Fatal("oe called")
	}
	coe(oe)
	if !ceCalled {
		t.Fatal("ce not called")
	}
}

func TestErrNilNil(t *testing.T) {
	var c1 Err
	c2 := c1.Nil()
	err := c2()
	assert.NoError(t, err)
}

func TestErrNilNotNilWithError(t *testing.T) {
	var c1 Err
	c2 := c1.Nil()
	called := false
	c1 = func() error {
		called = true
		return errors.New("error")
	}
	err := c2()
	assert.Error(t, err)
	assert.True(t, called)
}

func TestErrNilNotNilNoError(t *testing.T) {
	var c1 Err
	c2 := c1.Nil()
	called := false
	c1 = func() error {
		called = true
		return nil
	}
	err := c2()
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestOnErrWrap(t *testing.T) {
	oe := OnErr(func(err error) {
		msg := err.Error()
		expectedMsg := "test: error"
		assert.Equal(t, msg, expectedMsg)
	})
	oe = oe.Wrap("test")
	oe(errors.New("error"))
}

func TestWithOnErrWrap(t *testing.T) {
	coe := WithOnErr(func(oe OnErr) {
		oe(errors.New("error"))
	})
	coe = coe.Wrap("test")
	oeCalled := false
	oe := func(err error) {
		oeCalled = true
		msg := err.Error()
		expectedMsg := "test: error"
		assert.Equal(t, msg, expectedMsg)
	}
	coe(oe)
	assert.True(t, oeCalled)
}

func TestWithOnErrNilNil(t *testing.T) {
	var c1 WithOnErr
	c2 := c1.Nil()
	c2(func(err error) {})
}

func TestWithOnErrNilNotNil(t *testing.T) {
	var c1 WithOnErr
	c2 := c1.Nil()
	called := false
	c1 = func(oe OnErr) {
		called = true
	}
	c2(func(err error) {})
	assert.True(t, called)
}
