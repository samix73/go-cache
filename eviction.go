package cache

import "sync"

// EvictionStrategy defines the interface for eviction strategies used in the cache.
type EvictionStrategy[K comparable, V any] interface {
	// RecordAccess is called whenever a key is accessed in the cache.
	RecordAccess(key K)
	// RecordInsertion is called whenever a key is inserted into the cache.
	RecordInsertion(key K)
	// RecordDeletion is called whenever a key is deleted from the cache.
	RecordDeletion(key K)
	// Evict determines which key to evict from the cache based on the implemented strategy.
	// It returns the key to evict and a boolean indicating whether eviction should occur.
	Evict() (key K, shouldEvict bool)
	// Clear resets the eviction strategy's internal state.
	Clear()
}

var _ EvictionStrategy[any, any] = (*RandomEvictionStrategy[any, any])(nil)

// RandomEvictionStrategy is a simple eviction strategy that evicts a random key from the cache when necessary.
type RandomEvictionStrategy[K comparable, V any] struct {
	keys map[K]struct{}
	mu   sync.RWMutex
}

// NewRandomEvictionStrategy creates a new instance of the RandomEvictionStrategy.
func NewRandomEvictionStrategy[K comparable, V any]() *RandomEvictionStrategy[K, V] {
	return &RandomEvictionStrategy[K, V]{
		keys: make(map[K]struct{}),
	}
}

// RecordAccess is a no-op for the RandomEvictionStrategy since it does not track access patterns.
func (s *RandomEvictionStrategy[K, V]) RecordAccess(key K) {}

// RecordInsertion adds the key to the set of keys tracked by the eviction strategy.
func (s *RandomEvictionStrategy[K, V]) RecordInsertion(key K) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.keys[key] = struct{}{}
}

// RecordDeletion removes the key from the set of keys tracked by the eviction strategy.
func (s *RandomEvictionStrategy[K, V]) RecordDeletion(key K) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.keys, key)
}

// Evict randomly selects a key from the set of keys and returns it for eviction.
func (s *RandomEvictionStrategy[K, V]) Evict() (key K, shouldEvict bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for k := range s.keys {
		return k, true
	}

	var zero K
	return zero, false
}

// Clear resets the eviction strategy's internal state by clearing the set of keys.
func (s *RandomEvictionStrategy[K, V]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.keys = make(map[K]struct{})
}
