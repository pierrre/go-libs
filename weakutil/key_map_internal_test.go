package weakutil

import (
	"weak"
)

func KeyMapRawEntries[K any, V any](m *KeyMap[K, V]) (total, dead int) {
	m.m.Range(func(kp weak.Pointer[K], _ keyMapValue[V]) bool {
		total++
		_, alive := loadPointer(kp)
		if !alive {
			dead++
		}
		return true
	})
	return total, dead
}
