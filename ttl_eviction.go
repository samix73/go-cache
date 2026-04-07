package cache

import (
	"sync"
	"time"
)

var _ EvictionStrategy[any] = (*TTLEvictionStrategy[any])(nil)

// TTLEvictionStrategy implements a time-to-live eviction strategy for the cache.
// Entries are considered invalid after the configured TTL has elapsed since insertion.
type TTLEvictionStrategy[K comparable] struct {
	insertedAt map[K]time.Time
	mu         sync.RWMutex
	ttl        time.Duration
}

// NewTTLEvictionStrategy creates a new TTLEvictionStrategy with the given TTL.
func NewTTLEvictionStrategy[K comparable](ttl time.Duration) *TTLEvictionStrategy[K] {
	return &TTLEvictionStrategy[K]{
		insertedAt: make(map[K]time.Time),
		ttl:        ttl,
	}
}

func (t *TTLEvictionStrategy[K]) Evict() []K {
	t.mu.RLock()
	defer t.mu.RUnlock()

	now := time.Now()
	var expired []K
	for key, insertedAt := range t.insertedAt {
		if now.After(insertedAt.Add(t.ttl)) {
			expired = append(expired, key)
		}
	}

	return expired
}

func (t *TTLEvictionStrategy[K]) RecordAccess(keys ...K) {
	// TTL is based on insertion time, so access does not affect expiry.
}

func (t *TTLEvictionStrategy[K]) RecordInsertion(keys ...K) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	for _, key := range keys {
		t.insertedAt[key] = now
	}
}

func (t *TTLEvictionStrategy[K]) RecordDeletion(keys ...K) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, key := range keys {
		delete(t.insertedAt, key)
	}
}

func (t *TTLEvictionStrategy[K]) IsValid(key K) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	insertedAt, exists := t.insertedAt[key]
	if !exists {
		return false
	}

	return !time.Now().After(insertedAt.Add(t.ttl))
}

func (t *TTLEvictionStrategy[K]) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.insertedAt = make(map[K]time.Time)
}
