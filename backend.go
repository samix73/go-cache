package cache

import "context"

// Loader loads a value from a backing store for the given key.
// It returns the value, whether the key was found, and any error encountered.
// Implementations must be safe for concurrent use.
type Loader[K comparable, V any] interface {
	Load(ctx context.Context, key K) (V, bool, error)
}

// LoaderFunc is a function adapter that implements Loader.
type LoaderFunc[K comparable, V any] func(ctx context.Context, key K) (V, bool, error)

func (f LoaderFunc[K, V]) Load(ctx context.Context, key K) (V, bool, error) {
	return f(ctx, key)
}
