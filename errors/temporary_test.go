package errors

import (
	"fmt"
	"io"
	"testing"
)

func TestTemporaryTrue(t *testing.T) {
	err := newBase("error")
	err = WithTemporary(err, true)
	temporary := IsTemporary(err)
	if !temporary {
		t.Fatal("not temporary")
	}
}

func TestTemporaryFalse(t *testing.T) {
	err := newBase("error")
	err = WithTemporary(err, false)
	temporary := IsTemporary(err)
	if temporary {
		t.Fatal("temporary")
	}
}

func TestTemporaryDefault(t *testing.T) {
	err := newBase("error")
	temporary := IsTemporary(err)
	if !temporary {
		t.Fatal("not temporary")
	}
}

func TestTemporaryNil(t *testing.T) {
	err := WithTemporary(nil, true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTemporaryError(t *testing.T) {
	err := newBase("error")
	err = WithTemporary(err, true)
	s := err.Error()
	expected := "temporary true: error"
	if s != expected {
		t.Fatalf("unexpected message: got %q, want %q", s, expected)
	}
}

func TestTemporaryFormat(t *testing.T) {
	err := newBase("error")
	err = WithTemporary(err, true)
	s := fmt.Sprint(err) //nolint:gocritic // We want to test the Format method.
	expected := "temporary true: error"
	if s != expected {
		t.Fatalf("unexpected message: got %q, want %q", s, expected)
	}
}

func BenchmarkTemporaryFormat(b *testing.B) {
	err := newBase("error")
	err = WithTemporary(err, true)
	for i := 0; i < b.N; i++ {
		_, _ = fmt.Fprintf(io.Discard, "%+v", err)
	}
}
