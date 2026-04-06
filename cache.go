package cache

import "sync"

// Cache is a thread-safe in-memory cache that supports generic key-value pairs.
type Cache[K comparable, V any] struct {
	storage   map[K]V
	mu        sync.RWMutex
	copyOnSet CopyFunc[V]
	copyOnGet CopyFunc[V]
}

// NewCache creates a new instance of Cache.
func NewCache[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		storage: make(map[K]V),
		mu:      sync.RWMutex{},
	}
}

// Get retrieves the value associated with the given key.
// It returns the value and a boolean indicating whether the key exists in the cache.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, exists := c.storage[key]
	if exists && c.copyOnGet != nil {
		value = c.copyOnGet(value)
	}

	return value, exists
}

// Set adds or updates the value associated with the given key in the cache.
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.copyOnSet != nil {
		value = c.copyOnSet(value)
	}

	c.storage[key] = value
}

// Delete removes the value associated with the given key from the cache.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.storage, key)
}

// Clear removes all key-value pairs from the cache.
func (c *Cache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.storage = make(map[K]V)
}
