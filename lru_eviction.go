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

func (l *LRUEvictionStrategy[K, V]) Evict() (K, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.list.Len() == 0 {
		var zero K
		return zero, false
	}

	if l.list.Back() == nil {
		var zero K
		return zero, false
	}

	return l.list.Back().Value.(K), true
}

func (l *LRUEvictionStrategy[K, V]) RecordAccess(key K) {
	l.mu.Lock()
	defer l.mu.Unlock()

	element, exists := l.lookup[key]
	if !exists {
		return
	}

	l.list.MoveToFront(element)
}

func (l *LRUEvictionStrategy[K, V]) RecordInsertion(key K) {
	l.mu.Lock()
	defer l.mu.Unlock()

	element := l.list.PushFront(key)
	l.lookup[key] = element
}

func (l *LRUEvictionStrategy[K, V]) RecordDeletion(key K) {
	l.mu.Lock()
	defer l.mu.Unlock()

	element, exists := l.lookup[key]
	if !exists {
		return
	}

	l.list.Remove(element)
	delete(l.lookup, key)
}

func (l *LRUEvictionStrategy[K, V]) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.list = list.New()
	l.lookup = make(map[K]*list.Element)
}
