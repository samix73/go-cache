package cache

import (
	"container/list"
	"sync"
)

var _ EvictionStrategy[any, any] = (*LRUEvictionStrategy[any, any])(nil)

// LRUEvictionStrategy implements a least recently used eviction strategy for the cache.
type LRUEvictionStrategy[K comparable, V any] struct {
	list    *list.List
	lookup  map[K]*list.Element
	mu      sync.RWMutex
	maxSize uint
}

// NewLRUEvictionStrategy creates a new LRUEvictionStrategy with the given maximum size.
// When maxSize is 0, the strategy will never evict entries.
func NewLRUEvictionStrategy[K comparable, V any](maxSize uint) *LRUEvictionStrategy[K, V] {
	return &LRUEvictionStrategy[K, V]{
		list:    list.New(),
		lookup:  make(map[K]*list.Element),
		maxSize: maxSize,
	}
}

func (l *LRUEvictionStrategy[K, V]) Evict() []K {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.maxSize == 0 || uint(l.list.Len()) <= l.maxSize {
		return nil
	}

	overflow := uint(l.list.Len()) - l.maxSize
	keys := make([]K, 0, overflow)

	element := l.list.Back()
	for i := uint(0); i < overflow && element != nil; i++ {
		keys = append(keys, element.Value.(K))
		element = element.Prev()
	}

	return keys
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
