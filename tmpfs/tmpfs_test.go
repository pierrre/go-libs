package tmpfs

import (
	"os"
	"testing"
)

func TestDir(t *testing.T) {
	name, cl, err := Dir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer cl()
	f, err := os.Open(name) //nolint:gosec // This is not a security issue because we know what "name" contains.
	if err != nil {
		t.Fatal(err)
	}
	fi, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	d := fi.IsDir()
	if !d {
		t.Fatal("not dir")
	}
}

func TestFile(t *testing.T) {
	f, cl, err := File("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer cl()
	_, err = f.Stat()
	if err != nil {
		t.Fatal(err)
	}
}
