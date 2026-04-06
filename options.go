package cache

type CacheOptions[K comparable, V any] func(*Options[K, V])

type CopyFunc[V any] func(V) V

type Options[K comparable, V any] struct {
	copyOnSet CopyFunc[V]
	copyOnGet CopyFunc[V]
}

func WithCopyOnSet[K comparable, V any](copyFunc CopyFunc[V]) CacheOptions[K, V] {
	return func(o *Options[K, V]) {
		o.copyOnSet = copyFunc
	}
}

func WithCopyOnGet[K comparable, V any](copyFunc CopyFunc[V]) CacheOptions[K, V] {
	return func(o *Options[K, V]) {
		o.copyOnGet = copyFunc
	}
}

func applyOptions[K comparable, V any](opts ...CacheOptions[K, V]) *Options[K, V] {
	options := &Options[K, V]{}
	for _, opt := range opts {
		opt(options)
	}

	return options
}
