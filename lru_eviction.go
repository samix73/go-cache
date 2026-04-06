package cache

import (
	"container/list"
	"sync"
)

var _ EvictionStrategy[any, any] = (*LRUEvictionStrategy[any, any])(nil)

// LRUEvictionStrategy implements a least recently used eviction strategy for the cache.
type LRUEvictionStrategy[K comparable, V any] struct {
	list   *list.List
	lookup map[K]*list.Element
	mu     sync.RWMutex
}

func NewLRUEvictionStrategy[K comparable, V any]() *LRUEvictionStrategy[K, V] {
	return &LRUEvictionStrategy[K, V]{
		list:   list.New(),
		lookup: make(map[K]*list.Element),
	}
}

func (l *LRUEvictionStrategy[K, V]) Evict() []K {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.list.Len() == 0 {
		return nil
	}

	if l.list.Back() == nil {
		return nil
	}

	return []K{l.list.Back().Value.(K)}
}

func (l *LRUEvictionStrategy[K, V]) RecordAccess(keys ...K) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, key := range keys {
		element, exists := l.lookup[key]
		if !exists {
			continue
		}

		l.list.MoveToFront(element)
	}
}

func (l *LRUEvictionStrategy[K, V]) RecordInsertion(keys ...K) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, key := range keys {
		element := l.list.PushFront(key)
		l.lookup[key] = element
	}
}

func (l *LRUEvictionStrategy[K, V]) RecordDeletion(keys ...K) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, key := range keys {
		element, exists := l.lookup[key]
		if !exists {
			continue
		}

		l.list.Remove(element)
		delete(l.lookup, key)
	}
}

func (l *LRUEvictionStrategy[K, V]) IsValid(key K) bool {
	return true // LRU strategy does not invalidate keys based on access patterns, so we always return true.
}

func (l *LRUEvictionStrategy[K, V]) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.list = list.New()
	l.lookup = make(map[K]*list.Element)
}
