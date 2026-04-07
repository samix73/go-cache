package cache

import (
	"maps"
	"slices"
)

var _ EvictionStrategy[any] = (*CompositeEvictionStrategy[any])(nil)

type CompositeEvictionStrategy[K comparable] struct {
	strategies []EvictionStrategy[K]
}

func NewCompositeEvictionStrategy[K comparable](strategies ...EvictionStrategy[K]) *CompositeEvictionStrategy[K] {
	filtered := make([]EvictionStrategy[K], 0, len(strategies))
	for _, strategy := range strategies {
		if strategy == nil {
			continue
		}
		filtered = append(filtered, strategy)
	}

	return &CompositeEvictionStrategy[K]{
		strategies: filtered,
	}
}

func (c *CompositeEvictionStrategy[K]) Evict() []K {
	keysToEvict := make(map[K]struct{}, 0)
	for _, strategy := range c.strategies {
		keys := strategy.Evict()
		for _, key := range keys {
			keysToEvict[key] = struct{}{}
		}
	}

	return slices.Collect(maps.Keys(keysToEvict))
}

func (c *CompositeEvictionStrategy[K]) RecordAccess(keys ...K) {
	for _, strategy := range c.strategies {
		strategy.RecordAccess(keys...)
	}
}

func (c *CompositeEvictionStrategy[K]) RecordDeletion(keys ...K) {
	for _, strategy := range c.strategies {
		strategy.RecordDeletion(keys...)
	}
}

func (c *CompositeEvictionStrategy[K]) RecordInsertion(keys ...K) {
	for _, strategy := range c.strategies {
		strategy.RecordInsertion(keys...)
	}
}

func (c *CompositeEvictionStrategy[K]) IsValid(k K) bool {
	for _, strategy := range c.strategies {
		if !strategy.IsValid(k) {
			return false
		}
	}

	return true
}

func (c *CompositeEvictionStrategy[K]) Clear() {
	for _, strategy := range c.strategies {
		strategy.Clear()
	}
}
