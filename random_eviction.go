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

func (s *RandomEvictionStrategy[K, V]) RecordAccess(keys ...K) {}

func (s *RandomEvictionStrategy[K, V]) RecordInsertion(keys ...K) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, key := range keys {
		s.keys[key] = struct{}{}
	}
}

func (s *RandomEvictionStrategy[K, V]) RecordDeletion(keys ...K) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, key := range keys {
		delete(s.keys, key)
	}
}

func (s *RandomEvictionStrategy[K, V]) Evict() []K {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for k := range s.keys {
		return []K{k}
	}

	return nil
}

func (s *RandomEvictionStrategy[K, V]) IsValid(key K) bool {
	return true // Random eviction does not consider any key invalid, so it always returns true.
}

func (s *RandomEvictionStrategy[K, V]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.keys = make(map[K]struct{})
}
