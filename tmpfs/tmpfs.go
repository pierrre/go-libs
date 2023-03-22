// Package tmpfs provides function to access the temporary filesystem.
//
// The functions are simple wrapper for the os.*Temp functions.
// They return close function that guarantee that the files are properly closed and deleted.
package tmpfs

import (
	"fmt"
	"os"
)

// Dir is a helper for os.MkdirTemp.
//
// The returned close function deletes the directory.
func Dir(dir string, prefix string) (name string, cl func(), err error) {
	name, err = os.MkdirTemp(dir, prefix)
	if err != nil {
		return "", nil, fmt.Errorf("create: %w", err)
	}
	cl = func() {
		_ = os.RemoveAll(name)
	}
	return name, cl, nil
}

// File is a helper for os.CreateTemp.
//
// The returned close function closes and deletes the file.
func File(dir string, pattern string) (f *os.File, cl func(), err error) {
	f, err = os.CreateTemp(dir, pattern)
	if err != nil {
		return nil, nil, fmt.Errorf("create: %w", err)
	}
	cl = func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}
	return f, cl, nil
}
