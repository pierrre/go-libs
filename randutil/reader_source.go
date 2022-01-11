package randutil

import (
	"encoding/binary"
	"io"
)

// ReaderSource is a math/rand.Source implementation that reads data from a io.Reader.
//
// It can be used with crypto/rand.Reader.
type ReaderSource struct {
	Reader io.Reader
}

// Int63 implements math/rand.Source.
//
// It panics if the reader returns an error.
func (rs *ReaderSource) Int63() int64 {
	ui := rs.Uint64()
	ui &= 1<<63 - 1 // mask off sign bit to ensure positive number
	return int64(ui)
}

// Uint64 implements math/rand.Source64.
//
// It panics if the reader returns an error.
func (rs *ReaderSource) Uint64() uint64 {
	var b [8]byte
	_, err := io.ReadFull(rs.Reader, b[:])
	if err != nil {
		panic(err)
	}
	ui := binary.LittleEndian.Uint64(b[:])
	return ui
}

// Seed implements math/rand.Source.
//
// It's not supported and panics.
func (rs *ReaderSource) Seed(seed int64) {
	panic("seed is not supported")
}
