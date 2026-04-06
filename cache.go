package cache

import "sync"

// Cache is a thread-safe in-memory cache that supports generic key-value pairs.
type Cache[K comparable, V any] struct {
	storage map[K]V
	mu      sync.RWMutex
	options *Options[K, V]
}

// NewCache creates a new instance of Cache.
func NewCache[K comparable, V any](opts ...CacheOptions[K, V]) *Cache[K, V] {
	options := applyOptions(opts...)

	return &Cache[K, V]{
		storage: make(map[K]V),
		mu:      sync.RWMutex{},
		options: options,
	}
}

// Get retrieves the value associated with the given key.
// It returns the value and a boolean indicating whether the key exists in the cache.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, exists := c.storage[key]
	if exists && c.options.copyOnGet != nil {
		value = c.options.copyOnGet(value)
	}

	return value, exists
}

// Set adds or updates the value associated with the given key in the cache.
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.options.copyOnSet != nil {
		value = c.options.copyOnSet(value)
	}

	c.storage[key] = value
}

// Delete removes the value associated with the given key from the cache.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.storage, key)
}

// CompareAndSwap updates the value associated with the given key if the compareFn returns true.
// If the key does not exist, current will be the zero value of V.
func (c *Cache[K, V]) CompareAndSwap(key K, value V, compareFn func(current, new V) bool) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	currentValue := c.storage[key]
	if !compareFn(currentValue, value) {
		return false
	}

	if c.options.copyOnSet != nil {
		value = c.options.copyOnSet(value)
	}

	c.storage[key] = value

	return true
}

// Clear removes all key-value pairs from the cache.
func (c *Cache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.storage = make(map[K]V)
}
