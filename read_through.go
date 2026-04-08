package cache

import (
	"context"

	"golang.org/x/sync/singleflight"
)

// ReadThroughCache wraps a Cache and a Loader to provide read-through semantics.
// On a cache miss, the Loader is called to fetch the value from a backing store.
// Concurrent misses for the same key result in only one Loader call; all waiting
// callers receive the same value (singleflight / thundering-herd protection).
type ReadThroughCache[K comparable, V any] struct {
	cache  *Cache[K, V]
	loader Loader[K, V]
	group  singleflight.Group
}

// NewReadThroughCache creates a ReadThroughCache that wraps c and uses loader for
// cache misses. The underlying Cache and its eviction strategy are left unchanged.
func NewReadThroughCache[K comparable, V any](c *Cache[K, V], loader Loader[K, V]) *ReadThroughCache[K, V] {
	return &ReadThroughCache[K, V]{
		cache:  c,
		loader: loader,
	}
}

// Get returns the value for key.
//
//   - If the key is present in the in-memory cache the loader is not called.
//   - On a miss the loader is called exactly once per key, even when many
//     goroutines request the same missing key simultaneously.
//   - If the loader finds the value it is stored in the cache before returning.
//   - If the loader reports not-found the zero value is returned with found=false
//     and the result is not cached.
//   - If the loader returns an error the zero value is returned with the error
//     and the result is not cached.
func (r *ReadThroughCache[K, V]) Get(ctx context.Context, key K) (V, bool, error) {
	if v, ok := r.cache.Get(key); ok {
		return v, true, nil
	}

	// singleflight requires a string key; convert the generic key via a helper.
	sfKey := singleflightKey(key)

	type result struct {
		value V
		found bool
	}

	ch := r.group.DoChan(sfKey, func() (any, error) {
		v, found, err := r.loader.Load(ctx, key)
		if err != nil {
			return result{}, err
		}
		if found {
			r.cache.Set(key, v)
		}
		return result{value: v, found: found}, nil
	})

	select {
	case <-ctx.Done():
		var zero V
		return zero, false, ctx.Err()
	case res := <-ch:
		if res.Err != nil {
			var zero V
			return zero, false, res.Err
		}
		got := res.Val.(result)
		return got.value, got.found, nil
	}
}

// Set stores a key-value pair directly in the underlying cache.
// This is a convenience pass-through; it does not interact with the loader.
func (r *ReadThroughCache[K, V]) Set(key K, value V) {
	r.cache.Set(key, value)
}

// Delete removes a key from the underlying cache.
func (r *ReadThroughCache[K, V]) Delete(key K) {
	r.cache.Delete(key)
}
