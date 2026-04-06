package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrBatchSizeExceedsMaxSize = errors.New("batch size exceeds maximum cache size")
)

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

// StartEvictionRoutine starts a background goroutine that periodically checks the cache size and evicts entries if necessary based on the configured eviction strategy.
// The eviction routine will run until the provided context is canceled. This method will return an error if no eviction strategy is configured for the cache.
func (c *Cache[K, V]) StartEvictionRoutine(ctx context.Context, interval time.Duration) error {
	if c.options.evictionStrategy == nil {
		return errors.New("eviction routine requires an eviction strategy to be configured")
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	c.startEvictionRoutine(ctx, ticker.C)

	return nil
}

func (c *Cache[K, V]) startEvictionRoutine(ctx context.Context, ticker <-chan time.Time) {
	for {
		select {
		case <-ticker:
			c.mu.Lock()
			c.evict()
			c.mu.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

// delete removes the key-value pair associated with the given key from the cache without acquiring locks.
func (c *Cache[K, V]) delete(keys ...K) {
	if len(keys) == 0 {
		return
	}

	if c.options.evictionStrategy != nil {
		c.options.evictionStrategy.RecordDeletion(keys...)
	}
	for _, key := range keys {
		delete(c.storage, key)
	}
}

// evict checks if the cache has exceeded its maximum size and evicts an entry if necessary.
func (c *Cache[K, V]) evict() {
	if c.options.evictionStrategy == nil {
		return
	}

	if c.options.maxSize == 0 || uint(len(c.storage)) < c.options.maxSize {
		return
	}

	keysToDelete := c.options.evictionStrategy.Evict()
	c.delete(keysToDelete...)
}

func (c *Cache[K, V]) get(keys []K) map[K]V {
	results := make(map[K]V, len(keys))
	foundKeys := make([]K, 0, len(keys))
	for _, key := range keys {
		value, exists := c.storage[key]
		if !exists {
			continue
		}

		if c.options.evictionStrategy != nil && !c.options.evictionStrategy.IsValid(key) {
			continue
		}

		if c.options.copyOnGet != nil {
			value = c.options.copyOnGet(value)
		}

		results[key] = value
		foundKeys = append(foundKeys, key)
	}

	if c.options.evictionStrategy != nil {
		c.options.evictionStrategy.RecordAccess(foundKeys...)
	}

	return results
}

// Get retrieves the value associated with the given key.
// It returns the value and a boolean indicating whether the key exists in the cache.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	vals := c.get([]K{key})
	if len(vals) == 0 {
		var zeroValue V
		return zeroValue, false
	}

	val, exists := vals[key]
	if !exists {
		var zeroValue V
		return zeroValue, false
	}

	return val, true
}

// MGet retrieves the values associated with the given keys.
// It returns a map of keys to values for the keys that exist in the cache.
func (c *Cache[K, V]) MGet(keys ...K) map[K]V {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.get(keys)
}

// set adds or updates the value associated with the given key in the cache without acquiring locks.
func (c *Cache[K, V]) set(pairs map[K]V) error {
	if c.options.maxSize > 0 && len(pairs) > int(c.options.maxSize) {
		return ErrBatchSizeExceedsMaxSize
	}

	for key, value := range pairs {
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

	return nil
}

// Set adds or updates the value associated with the given key in the cache.
func (c *Cache[K, V]) Set(key K, value V) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.set(map[K]V{key: value})
}

// MSet adds or updates the values associated with the given keys in the cache.
func (c *Cache[K, V]) MSet(pairs map[K]V) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.set(pairs)
}

// Delete removes the value associated with the given key from the cache.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.delete(key)
}

// CompareAndSwap updates the value associated with the given key if the compareFn returns true.
func (c *Cache[K, V]) CompareAndSwap(key K, value V, compareFn func(current, new V) bool) (bool, error) {
	if compareFn == nil {
		return false, errors.New("compare function cannot be nil")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	currentValue, exists := c.storage[key]
	if !exists || !compareFn(currentValue, value) {
		return false, nil
	}

	if err := c.set(map[K]V{key: value}); err != nil {
		return false, err
	}

	return true, nil
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
