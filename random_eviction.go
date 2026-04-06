package cache

import "sync"

var _ EvictionStrategy[any, any] = (*RandomEvictionStrategy[any, any])(nil)

// RandomEvictionStrategy is a simple eviction strategy that evicts a random key from the cache when necessary.
type RandomEvictionStrategy[K comparable, V any] struct {
	keys map[K]struct{}
	mu   sync.RWMutex
}

func NewRandomEvictionStrategy[K comparable, V any]() *RandomEvictionStrategy[K, V] {
	return &RandomEvictionStrategy[K, V]{
		keys: make(map[K]struct{}),
	}
}

func (s *RandomEvictionStrategy[K, V]) RecordAccess(key K) {}

func (s *RandomEvictionStrategy[K, V]) RecordInsertion(key K) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.keys[key] = struct{}{}
}

func (s *RandomEvictionStrategy[K, V]) RecordDeletion(key K) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.keys, key)
}

func (s *RandomEvictionStrategy[K, V]) Evict() (key K, shouldEvict bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for k := range s.keys {
		return k, true
	}

	var zero K
	return zero, false
}

func (s *RandomEvictionStrategy[K, V]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.keys = make(map[K]struct{})
}
