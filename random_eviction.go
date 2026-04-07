package cache

import (
	"math/rand/v2"
	"sync"
)

var _ EvictionStrategy[any] = (*RandomEvictionStrategy[any])(nil)

// RandomEvictionStrategy implements a random replacement eviction strategy for the cache.
// When the cache exceeds the configured maximum size, random entries are chosen for eviction.
// RecordAccess is a no-op, making it lockless on the read path.
type RandomEvictionStrategy[K comparable] struct {
	keys    []K
	index   map[K]int
	mu      sync.RWMutex
	maxSize uint
}

// NewRandomEvictionStrategy creates a new RandomEvictionStrategy with the given maximum size.
// When maxSize is 0, the strategy will never evict entries.
func NewRandomEvictionStrategy[K comparable](maxSize uint) *RandomEvictionStrategy[K] {
	return &RandomEvictionStrategy[K]{
		keys:    make([]K, 0),
		index:   make(map[K]int),
		maxSize: maxSize,
	}
}

func (r *RandomEvictionStrategy[K]) Evict() []K {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.maxSize == 0 || uint(len(r.keys)) <= r.maxSize {
		return nil
	}

	overflow := uint(len(r.keys)) - r.maxSize

	// Use a partial Fisher-Yates shuffle on a copy of the index positions to
	// select exactly overflow distinct random keys without modifying internal state.
	positions := make([]int, len(r.keys))
	for i := range positions {
		positions[i] = i
	}

	result := make([]K, overflow)
	for i := uint(0); i < overflow; i++ {
		j := i + uint(rand.N(uint(len(positions))-i))
		positions[i], positions[j] = positions[j], positions[i]
		result[i] = r.keys[positions[i]]
	}

	return result
}

func (r *RandomEvictionStrategy[K]) RecordAccess(keys ...K) {
	// Random replacement does not consider access patterns.
}

func (r *RandomEvictionStrategy[K]) RecordInsertion(keys ...K) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, key := range keys {
		if _, exists := r.index[key]; exists {
			continue
		}
		r.index[key] = len(r.keys)
		r.keys = append(r.keys, key)
	}
}

func (r *RandomEvictionStrategy[K]) RecordDeletion(keys ...K) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, key := range keys {
		idx, exists := r.index[key]
		if !exists {
			continue
		}

		// Swap the target with the last element and shrink the slice (O(1) delete).
		last := len(r.keys) - 1
		if idx != last {
			lastKey := r.keys[last]
			r.keys[idx] = lastKey
			r.index[lastKey] = idx
		}
		r.keys = r.keys[:last]
		delete(r.index, key)
	}
}

func (r *RandomEvictionStrategy[K]) IsValid(key K) bool {
	return true // Random replacement does not invalidate keys based on access patterns.
}

func (r *RandomEvictionStrategy[K]) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.keys = make([]K, 0)
	r.index = make(map[K]int)
}
