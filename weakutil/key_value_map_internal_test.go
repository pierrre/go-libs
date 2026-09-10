package weakutil

import (
	"weak"
)

func KeyValueMapRawEntry[K any, V any](m *KeyValueMap[K, V], kp weak.Pointer[K]) (present, alive bool) {
	mv, ok := m.m.Load(kp)
	if !ok {
		return false, false
	}
	_, alive = loadPointer(mv.value)
	return true, alive
}
