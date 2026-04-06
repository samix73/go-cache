package cache

import "sync"

// Cache is a thread-safe in-memory cache that supports generic key-value pairs.
type Cache[K comparable, V any] struct {
	storage map[K]V
	mu      sync.RWMutex
	options Options[K, V]
}

// NewCache creates a new instance of Cache.
func NewCache[K comparable, V any](opts ...CacheOptions[K, V]) *Cache[K, V] {
	return &Cache[K, V]{
		storage: make(map[K]V),

		options: applyOptions(opts...),
	}
}

// delete removes the key-value pair associated with the given key from the cache without acquiring locks.
func (c *Cache[K, V]) delete(key K) {
	if c.options.evictionStrategy != nil {
		c.options.evictionStrategy.RecordDeletion(key)
	}
	delete(c.storage, key)
}

// set adds or updates the value associated with the given key in the cache without acquiring locks.
func (c *Cache[K, V]) set(key K, value V) {
	if c.options.copyOnSet != nil {
		value = c.options.copyOnSet(value)
	}

	_, exists := c.storage[key]
	if !exists {
		c.evict()
		if c.options.evictionStrategy != nil {
			c.options.evictionStrategy.RecordInsertion(key)
		}
	} else if c.options.evictionStrategy != nil {
		// Record access for existing keys to update their status in the eviction strategy.
		c.options.evictionStrategy.RecordAccess(key)
	}

	c.storage[key] = value
}

// evict checks if the cache has exceeded its maximum size and evicts an entry if necessary.
func (c *Cache[K, V]) evict() {
	if c.options.evictionStrategy == nil {
		return
	}

	if c.options.maxSize == 0 || uint(len(c.storage)) < c.options.maxSize {
		return
	}

	key, shouldEvict := c.options.evictionStrategy.Evict()
	if shouldEvict {
		c.delete(key)
	}
}

// Get retrieves the value associated with the given key.
// It returns the value and a boolean indicating whether the key exists in the cache.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, exists := c.storage[key]
	if !exists {
		var zeroValue V
		return zeroValue, false
	}

	if c.options.copyOnGet != nil {
		value = c.options.copyOnGet(value)
	}

	if c.options.evictionStrategy != nil {
		c.options.evictionStrategy.RecordAccess(key)
	}

	return value, exists
}

// Set adds or updates the value associated with the given key in the cache.
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.set(key, value)
}

// Delete removes the value associated with the given key from the cache.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.delete(key)
}

// CompareAndSwap updates the value associated with the given key if the compareFn returns true.
func (c *Cache[K, V]) CompareAndSwap(key K, value V, compareFn func(current, new V) bool) bool {
	if compareFn == nil {
		return false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	currentValue, exists := c.storage[key]
	if !exists || !compareFn(currentValue, value) {
		return false
	}

	c.set(key, value)

	return true
}

// Clear removes all key-value pairs from the cache.
func (c *Cache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.storage = make(map[K]V)
	if c.options.evictionStrategy != nil {
		c.options.evictionStrategy.Clear()
	}
}
