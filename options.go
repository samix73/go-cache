package cache

type CacheOptions[K comparable, V any] func(*Cache[K, V])

type CopyFunc[V any] func(V) V

func WithCopyOnSet[K comparable, V any](copyFunc CopyFunc[V]) CacheOptions[K, V] {
	return func(c *Cache[K, V]) {
		c.copyOnSet = copyFunc
	}
}

func WithCopyOnGet[K comparable, V any](copyFunc CopyFunc[V]) CacheOptions[K, V] {
	return func(c *Cache[K, V]) {
		c.copyOnGet = copyFunc
	}
}
