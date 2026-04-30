package cache

type CacheOption[K comparable, V any] func(*cacheConfig[K, V])

type CopyFunc[V any] func(V) V

type cacheConfig[K comparable, V any] struct {
	copyOnSet            CopyFunc[V]
	copyOnGet            CopyFunc[V]
	evictionStrategy     EvictionStrategy[K]
	disableEvictionOnSet bool
}

// WithCopyOnSet sets a function that will be called to create a copy of the value when it is added to the cache.
func WithCopyOnSet[K comparable, V any](copyFunc CopyFunc[V]) CacheOption[K, V] {
	return func(o *cacheConfig[K, V]) {
		o.copyOnSet = copyFunc
	}
}

// WithCopyOnGet sets a function that will be called to create a copy of the value when it is retrieved from the cache.
func WithCopyOnGet[K comparable, V any](copyFunc CopyFunc[V]) CacheOption[K, V] {
	return func(o *cacheConfig[K, V]) {
		o.copyOnGet = copyFunc
	}
}

// WithEvictionStrategy sets the eviction strategy to use for the cache.
func WithEvictionStrategy[K comparable, V any](strategy EvictionStrategy[K]) CacheOption[K, V] {
	return func(o *cacheConfig[K, V]) {
		o.evictionStrategy = strategy
	}
}

// WithDisableEvictionOnSet disables the eviction check that normally runs when a new key is inserted via Set or MSet.
// When this option is set, eviction will only happen via the background eviction routine started with StartEvictionRoutine.
func WithDisableEvictionOnSet[K comparable, V any]() CacheOption[K, V] {
	return func(o *cacheConfig[K, V]) {
		o.disableEvictionOnSet = true
	}
}

func applyCacheOptions[K comparable, V any](opts ...CacheOption[K, V]) cacheConfig[K, V] {
	var options cacheConfig[K, V]
	for _, opt := range opts {
		opt(&options)
	}

	return options
}
