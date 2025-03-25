package tmpfs

import (
	"os"
	"testing"

	"github.com/pierrre/assert"
)

func TestDir(t *testing.T) {
	name, cl, err := Dir("", "")
	assert.NoError(t, err)
	defer cl()
	f, err := os.Open(name) //nolint:gosec // We want to open a file.
	assert.NoError(t, err)
	fi, err := f.Stat()
	assert.NoError(t, err)
	d := fi.IsDir()
	assert.True(t, d)
}

func TestFile(t *testing.T) {
	f, cl, err := File("", "")
	assert.NoError(t, err)
	defer cl()
	_, err = f.Stat()
	assert.NoError(t, err)
}
