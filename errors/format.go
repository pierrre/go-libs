package errors

import (
	"fmt"
	"io"

	"github.com/pierrre/go-libs/bufpool"
)

var bufPool = &bufpool.Pool{}

// Formattable represents a formattable error.
type Formattable interface {
	error
	WriteErrorMessage(w io.Writer, verbose bool) bool
}

// Format formats an error.
func Format(err Formattable, s fmt.State, verb rune) {
	switch {
	case verb == 'v' && s.Flag('+'):
		writeError(s, err, true)
	case verb == 'v' || verb == 's':
		writeError(s, err, false)
	case verb == 'q':
		_, _ = fmt.Fprintf(s, "%q", Error(err))
	}
}

// Error formats an error on a single line.
func Error(err Formattable) string {
	buf := bufPool.Get()
	defer bufPool.Put(buf)
	writeError(buf, err, false)
	return buf.String()
}

func writeError(w io.Writer, err Formattable, verbose bool) {
	var separator string
	if verbose {
		separator = "\n"
	} else {
		separator = ": "
	}
	for {
		ok := err.WriteErrorMessage(w, verbose)
		werr := Unwrap(err)
		if werr == nil {
			return
		}
		if ok {
			_, _ = io.WriteString(w, separator)
		}
		err, ok = werr.(Formattable) //nolint:errorlint // We need to check if the current error value implements the interface.
		if !ok {
			_, _ = io.WriteString(w, werr.Error())
			return
		}
	}
}
