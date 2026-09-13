package weakutil

import (
	"weak"
)

func loadPointer[T any](p weak.Pointer[T]) (*T, bool) {
	if p == (weak.Pointer[T]{}) {
		return nil, true
	}
	v := p.Value()
	return v, v != nil
}
