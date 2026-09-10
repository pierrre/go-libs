package weakutil

func ValueMapRawEntry[K comparable, V any](m *ValueMap[K, V], key K) (present, alive bool) {
	mv, ok := m.m.Load(key)
	if !ok {
		return false, false
	}
	_, alive = loadPointer(mv.value)
	return true, alive
}
