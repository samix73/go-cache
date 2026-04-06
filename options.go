package cache

type CacheOptions[K comparable, V any] func(*Options[K, V])

type CopyFunc[V any] func(V) V

type Options[K comparable, V any] struct {
	copyOnSet        CopyFunc[V]
	copyOnGet        CopyFunc[V]
	maxSize          uint
	evictionStrategy EvictionStrategy[K, V]
}

// WithCopyOnSet sets a function that will be called to create a copy of the value when it is added to the cache.
func WithCopyOnSet[K comparable, V any](copyFunc CopyFunc[V]) CacheOptions[K, V] {
	return func(o *Options[K, V]) {
		o.copyOnSet = copyFunc
	}
}

// WithCopyOnGet sets a function that will be called to create a copy of the value when it is retrieved from the cache.
func WithCopyOnGet[K comparable, V any](copyFunc CopyFunc[V]) CacheOptions[K, V] {
	return func(o *Options[K, V]) {
		o.copyOnGet = copyFunc
	}
}

// WithEvictionStrategy sets the eviction strategy to use for the cache.
// maxSize specifies the cache capacity, and strategy determines which entries are evicted when adding
// a new entry would cause the cache to exceed that limit.
func WithEvictionStrategy[K comparable, V any](maxSize uint, strategy EvictionStrategy[K, V]) CacheOptions[K, V] {
	return func(o *Options[K, V]) {
		o.maxSize = maxSize
		o.evictionStrategy = strategy
	}
}

func applyOptions[K comparable, V any](opts ...CacheOptions[K, V]) Options[K, V] {
	var options Options[K, V]
	for _, opt := range opts {
		opt(&options)
	}

	return options
}
