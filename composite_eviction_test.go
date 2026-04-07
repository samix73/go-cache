package cache

import "testing"

type mockEvictionStrategy[K comparable, V any] struct {
	evictKeys      []K
	validity       map[K]bool
	recordAccess   [][]K
	recordInsert   [][]K
	recordDelete   [][]K
	clearCallCount int
}

func (m *mockEvictionStrategy[K, V]) Evict() []K {
	out := make([]K, len(m.evictKeys))
	copy(out, m.evictKeys)
	return out
}

func (m *mockEvictionStrategy[K, V]) RecordAccess(keys ...K) {
	copied := make([]K, len(keys))
	copy(copied, keys)
	m.recordAccess = append(m.recordAccess, copied)
}

func (m *mockEvictionStrategy[K, V]) RecordDeletion(keys ...K) {
	copied := make([]K, len(keys))
	copy(copied, keys)
	m.recordDelete = append(m.recordDelete, copied)
}

func (m *mockEvictionStrategy[K, V]) RecordInsertion(keys ...K) {
	copied := make([]K, len(keys))
	copy(copied, keys)
	m.recordInsert = append(m.recordInsert, copied)
}

func (m *mockEvictionStrategy[K, V]) IsValid(key K) bool {
	if m.validity == nil {
		return true
	}
	valid, exists := m.validity[key]
	if !exists {
		return true
	}
	return valid
}

func (m *mockEvictionStrategy[K, V]) Clear() {
	m.clearCallCount++
}

func TestNewCompositeEvictionStrategyFiltersNil(t *testing.T) {
	t.Parallel()

	left := &mockEvictionStrategy[string, int]{}
	right := &mockEvictionStrategy[string, int]{}

	composite := NewCompositeEvictionStrategy[string, int](left, nil, right, nil)
	if len(composite.strategies) != 2 {
		t.Fatalf("expected 2 non-nil strategies, got %d", len(composite.strategies))
	}
}

func TestCompositeEvictionStrategyEvictDeduplicatesKeys(t *testing.T) {
	t.Parallel()

	first := &mockEvictionStrategy[string, int]{evictKeys: []string{"a", "b", "c"}}
	second := &mockEvictionStrategy[string, int]{evictKeys: []string{"b", "d"}}
	third := &mockEvictionStrategy[string, int]{evictKeys: []string{"a"}}

	composite := NewCompositeEvictionStrategy[string, int](first, second, third)
	got := composite.Evict()

	if len(got) != 4 {
		t.Fatalf("expected 4 unique keys, got %d: %v", len(got), got)
	}

	seen := make(map[string]struct{}, len(got))
	for _, key := range got {
		seen[key] = struct{}{}
	}
	for _, key := range []string{"a", "b", "c", "d"} {
		if _, exists := seen[key]; !exists {
			t.Fatalf("expected eviction key %q to be present; got %v", key, got)
		}
	}
}

func TestCompositeEvictionStrategyForwardsHookCalls(t *testing.T) {
	t.Parallel()

	first := &mockEvictionStrategy[string, int]{}
	second := &mockEvictionStrategy[string, int]{}
	composite := NewCompositeEvictionStrategy[string, int](first, second)

	composite.RecordAccess("a", "b")
	composite.RecordInsertion("c")
	composite.RecordDeletion("d")
	composite.Clear()

	for i, strategy := range []*mockEvictionStrategy[string, int]{first, second} {
		if len(strategy.recordAccess) != 1 {
			t.Fatalf("strategy %d expected 1 access call, got %d", i, len(strategy.recordAccess))
		}
		if len(strategy.recordInsert) != 1 {
			t.Fatalf("strategy %d expected 1 insertion call, got %d", i, len(strategy.recordInsert))
		}
		if len(strategy.recordDelete) != 1 {
			t.Fatalf("strategy %d expected 1 deletion call, got %d", i, len(strategy.recordDelete))
		}
		if strategy.clearCallCount != 1 {
			t.Fatalf("strategy %d expected 1 clear call, got %d", i, strategy.clearCallCount)
		}
	}
}

func TestCompositeEvictionStrategyIsValidRequiresAllStrategies(t *testing.T) {
	t.Parallel()

	t.Run("all valid", func(t *testing.T) {
		t.Parallel()

		first := &mockEvictionStrategy[string, int]{validity: map[string]bool{"k": true}}
		second := &mockEvictionStrategy[string, int]{validity: map[string]bool{"k": true}}

		composite := NewCompositeEvictionStrategy[string, int](first, second)
		if !composite.IsValid("k") {
			t.Fatal("expected key to be valid when all strategies allow it")
		}
	})

	t.Run("one invalid", func(t *testing.T) {
		t.Parallel()

		first := &mockEvictionStrategy[string, int]{validity: map[string]bool{"k": true}}
		second := &mockEvictionStrategy[string, int]{validity: map[string]bool{"k": false}}

		composite := NewCompositeEvictionStrategy[string, int](first, second)
		if composite.IsValid("k") {
			t.Fatal("expected key to be invalid when any strategy rejects it")
		}
	})
}
