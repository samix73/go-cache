package cache

type CacheOptions[K comparable, V any] func(*Options[K, V])

type CopyFunc[V any] func(V) V

type Options[K comparable, V any] struct {
	copyOnSet              CopyFunc[V]
	copyOnGet              CopyFunc[V]
	evictionStrategy       EvictionStrategy[K, V]
	disableEvictionOnSet   bool
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
func WithEvictionStrategy[K comparable, V any](strategy EvictionStrategy[K, V]) CacheOptions[K, V] {
	return func(o *Options[K, V]) {
		o.evictionStrategy = strategy
	}
}

// WithDisableEvictionOnSet disables the eviction check that normally runs when a new key is inserted via Set or MSet.
// When this option is set, eviction will only happen via the background eviction routine started with StartEvictionRoutine.
func WithDisableEvictionOnSet[K comparable, V any]() CacheOptions[K, V] {
	return func(o *Options[K, V]) {
		o.disableEvictionOnSet = true
	}
}

func applyOptions[K comparable, V any](opts ...CacheOptions[K, V]) Options[K, V] {
	var options Options[K, V]
	for _, opt := range opts {
		opt(&options)
	}

	return options
}
