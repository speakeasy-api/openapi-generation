package generate

import "sync"

// SyncMap is a generic type-safe wrapper around sync.Map.
type SyncMap[K comparable, V any] struct {
	m sync.Map
}

// Store sets the value for a key.
func (s *SyncMap[K, V]) Store(key K, value V) {
	s.m.Store(key, value)
}

// Load returns the value stored in the map for a key, or the zero value if no value is present.
func (s *SyncMap[K, V]) Load(key K) (V, bool) {
	val, ok := s.m.Load(key)
	if !ok {
		var zero V
		return zero, false
	}
	return val.(V), true
}

// Delete removes the value for a key.
func (s *SyncMap[K, V]) Delete(key K) {
	s.m.Delete(key)
}

// Range calls f sequentially for each key and value present in the map.
func (s *SyncMap[K, V]) Range(f func(key K, value V) bool) {
	s.m.Range(func(key, value any) bool {
		return f(key.(K), value.(V))
	})
}
