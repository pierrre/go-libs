package errors

import (
	"fmt"
	"io"
	"testing"

	"github.com/pierrre/go-libs/internal/testutil"
)

func TestValue(t *testing.T) {
	err := newBase("error")
	err = WithValue(err, "foo", "bar")
	vals := Values(err)
	expected := map[string]interface{}{
		"foo": "bar",
	}
	testutil.Compare(t, "unexpected values", vals, expected)
}

func TestValueOverWrite(t *testing.T) {
	err := newBase("error")
	err = WithValue(err, "test", 1)
	err = WithValue(err, "test", 2)
	vals := Values(err)
	expected := map[string]interface{}{
		"test": 2,
	}
	testutil.Compare(t, "unexpected values", vals, expected)
}

func TestValueNil(t *testing.T) {
	err := WithValue(nil, "foo", "bar")
	if err != nil {
		t.Fatal(err)
	}
}

func TestValuesEmpty(t *testing.T) {
	err := newBase("error")
	vals := Values(err)
	if len(vals) != 0 {
		t.Fatalf("values not empty: got %#v", vals)
	}
}

func TestValueError(t *testing.T) {
	err := newBase("error")
	err = WithValue(err, "foo", "bar")
	s := err.Error()
	expected := "error"
	if s != expected {
		t.Fatalf("unexpected message: got %q, want %q", s, expected)
	}
}

func TestValueFormat(t *testing.T) {
	err := newBase("error")
	err = WithValue(err, "foo", "bar")
	s := fmt.Sprintf("%+v", err)
	expected := "value foo = bar\nerror"
	if s != expected {
		t.Fatalf("unexpected message: got %q, want %q", s, expected)
	}
}

func BenchmarkValueFormat(b *testing.B) {
	err := newBase("error")
	err = WithValue(err, "foo", "bar")
	for i := 0; i < b.N; i++ {
		_, _ = fmt.Fprintf(io.Discard, "%+v", err)
	}
}
