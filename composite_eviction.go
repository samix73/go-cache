package cache

import (
	"maps"
	"slices"
)

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

func (c *CompositeEvictionStrategy[K, V]) Evict() []K {
	keysToEvict := make(map[K]struct{}, 0)
	for _, strategy := range c.strategies {
		keys := strategy.Evict()
		for _, key := range keys {
			keysToEvict[key] = struct{}{}
		}
	}

	return slices.Collect(maps.Keys(keysToEvict))
}

func (c *CompositeEvictionStrategy[K, V]) RecordAccess(keys ...K) {
	for _, strategy := range c.strategies {
		strategy.RecordAccess(keys...)
	}
}

func (c *CompositeEvictionStrategy[K, V]) RecordDeletion(keys ...K) {
	for _, strategy := range c.strategies {
		strategy.RecordDeletion(keys...)
	}
}

func (c *CompositeEvictionStrategy[K, V]) RecordInsertion(keys ...K) {
	for _, strategy := range c.strategies {
		strategy.RecordInsertion(keys...)
	}
}

func (c *CompositeEvictionStrategy[K, V]) IsValid(k K) bool {
	for _, strategy := range c.strategies {
		if !strategy.IsValid(k) {
			return false
		}
	}

	return true
}

func (c *CompositeEvictionStrategy[K, V]) Clear() {
	for _, strategy := range c.strategies {
		strategy.Clear()
	}
}
