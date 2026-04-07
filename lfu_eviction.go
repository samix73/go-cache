package cache

import (
	"slices"
	"sync"
)

var _ EvictionStrategy[any] = (*LFUEvictionStrategy[any])(nil)

type LFUEvictionStrategy[K comparable] struct {
	maxSize int
	entries map[K]lfuEntry
	mu      sync.RWMutex
	n       int
}

type lfuEntry struct {
	frequency  int
	accessTime int
}

func NewLFUEvictionStrategy[K comparable](maxSize int) *LFUEvictionStrategy[K] {
	return &LFUEvictionStrategy[K]{
		maxSize: maxSize,
		entries: make(map[K]lfuEntry),
	}
}

func (l *LFUEvictionStrategy[K]) accessTime() int {
	n := l.n
	l.n++

	return n
}

func (l *LFUEvictionStrategy[K]) Evict() []K {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.maxSize <= 0 || len(l.entries) <= l.maxSize {
		return nil
	}

	offset := len(l.entries) - l.maxSize

	type canidate struct {
		key   K
		entry lfuEntry
	}

	candidates := make([]canidate, 0, len(l.entries))
	for key, entry := range l.entries {
		candidates = append(candidates, canidate{
			key:   key,
			entry: entry,
		})
	}

	// Sort candidates by frequency, then by access time to break ties (oldest first).
	slices.SortStableFunc(candidates, func(a, b canidate) int {
		if a.entry.frequency == b.entry.frequency {
			return a.entry.accessTime - b.entry.accessTime
		}

		return a.entry.frequency - b.entry.frequency
	})

	evictedKeys := make([]K, 0, offset)
	for i := range offset {
		evictedKeys = append(evictedKeys, candidates[i].key)
	}

	return evictedKeys
}

// IsValid always returns true for LFU strategy since it does not invalidate keys based on access patterns.
func (l *LFUEvictionStrategy[K]) IsValid(k K) bool {
	return true
}

// RecordAccess increments the frequency count for the accessed keys and updates their access time.
func (l *LFUEvictionStrategy[K]) RecordAccess(keys ...K) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, key := range keys {
		if entry, exists := l.entries[key]; exists {
			entry.frequency++
			entry.accessTime = l.accessTime()
			l.entries[key] = entry
		}
	}
}

func (l *LFUEvictionStrategy[K]) RecordDeletion(keys ...K) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, key := range keys {
		delete(l.entries, key)
	}
}

func (l *LFUEvictionStrategy[K]) RecordInsertion(keys ...K) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, key := range keys {
		l.entries[key] = lfuEntry{
			frequency:  1,
			accessTime: l.accessTime(),
		}
	}
}

func (l *LFUEvictionStrategy[K]) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.entries = make(map[K]lfuEntry)
	l.n = 0
}
