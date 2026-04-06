package cache

var _ EvictionStrategy[string, int] = (*CompositeEvictionStrategy[string, int])(nil)

type CompositeEvictionStrategy[K comparable, V any] struct {
	strategies []EvictionStrategy[K, V]
}

func NewCompositeEvictionStrategy[K comparable, V any](strategies ...EvictionStrategy[K, V]) *CompositeEvictionStrategy[K, V] {
	filtered := make([]EvictionStrategy[K, V], 0, len(strategies))
	for _, strategy := range strategies {
		if strategy == nil {
			continue
		}
		filtered = append(filtered, strategy)
	}

	return &CompositeEvictionStrategy[K, V]{
		strategies: filtered,
	}
}

func (c *CompositeEvictionStrategy[K, V]) Evict() (K, bool) {
	for _, strategy := range c.strategies {
		key, shouldEvict := strategy.Evict()
		if shouldEvict {
			return key, true
		}
	}

	var zeroKey K
	return zeroKey, false
}

func (c *CompositeEvictionStrategy[K, V]) RecordAccess(key K) {
	for _, strategy := range c.strategies {
		strategy.RecordAccess(key)
	}
}

func (c *CompositeEvictionStrategy[K, V]) RecordDeletion(key K) {
	for _, strategy := range c.strategies {
		strategy.RecordDeletion(key)
	}
}

func (c *CompositeEvictionStrategy[K, V]) RecordInsertion(key K) {
	for _, strategy := range c.strategies {
		strategy.RecordInsertion(key)
	}
}

func (c *CompositeEvictionStrategy[K, V]) Clear() {
	for _, strategy := range c.strategies {
		strategy.Clear()
	}
}
