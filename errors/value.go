package errors

import (
	"fmt"
	"io"
)

// ValueWriter writes a value attached to an error.
var ValueWriter = func(w io.Writer, v interface{}) {
	fmt.Fprint(w, v)
}

// WithValue adds a value to an error.
func WithValue(err error, key string, val interface{}) error {
	if err == nil {
		return nil
	}
	return &value{
		err:   err,
		key:   key,
		value: val,
	}
}

type value struct {
	err   error
	key   string
	value interface{}
}

func (err *value) Value() (key string, value interface{}) {
	return err.key, err.value
}

func (err *value) WriteErrorMessage(w io.Writer, verbose bool) bool {
	if !verbose {
		return false
	}
	_, _ = fmt.Fprintf(w, "value %s = ", err.key)
	ValueWriter(w, err.value)
	return true
}

func (err *value) Error() string                 { return Error(err) }
func (err *value) Format(s fmt.State, verb rune) { Format(err, s, verb) }
func (err *value) Unwrap() error                 { return err.err }

// Values returns the values associated to an error.
func Values(err error) map[string]interface{} {
	vals := make(map[string]interface{})
	for ; err != nil; err = Unwrap(err) {
		err, ok := err.(*value) //nolint:errorlint // We want to compare the current error.
		if !ok {
			continue
		}
		k, v := err.Value()
		_, ok = vals[k]
		if ok {
			continue
		}
		vals[k] = v
	}
	return vals
}
