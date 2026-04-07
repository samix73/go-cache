package cache

import "testing"

func TestLFUEvictionStrategyEvict(t *testing.T) {
	t.Parallel()

	t.Run("does not evict when under max size", func(t *testing.T) {
		t.Parallel()

		strategy := NewLFUEvictionStrategy[string](3)
		strategy.RecordInsertion("a", "b")

		if keys := strategy.Evict(); len(keys) != 0 {
			t.Fatalf("expected no eviction, got %v", keys)
		}
	})

	t.Run("does not evict when at max size", func(t *testing.T) {
		t.Parallel()

		strategy := NewLFUEvictionStrategy[string](3)
		strategy.RecordInsertion("a", "b", "c")

		if keys := strategy.Evict(); len(keys) != 0 {
			t.Fatalf("expected no eviction at exact max size, got %v", keys)
		}
	})

	t.Run("evicts least frequently used key", func(t *testing.T) {
		t.Parallel()

		strategy := NewLFUEvictionStrategy[string](2)
		strategy.RecordInsertion("a", "b", "c")

		strategy.RecordAccess("a", "a", "b")

		keys := strategy.Evict()
		if len(keys) != 1 {
			t.Fatalf("expected 1 key evicted, got %d: %v", len(keys), keys)
		}
		if keys[0] != "c" {
			t.Fatalf("expected key c to be evicted, got %v", keys)
		}
	})

	t.Run("evicts oldest on equal frequency", func(t *testing.T) {
		t.Parallel()

		strategy := NewLFUEvictionStrategy[string](1)
		strategy.RecordInsertion("a", "b")

		keys := strategy.Evict()
		if len(keys) != 1 {
			t.Fatalf("expected 1 key evicted, got %d: %v", len(keys), keys)
		}
		if keys[0] != "a" {
			t.Fatalf("expected oldest key a to be evicted on tie, got %v", keys)
		}
	})

	t.Run("evicts correct number of keys when over max size", func(t *testing.T) {
		t.Parallel()

		strategy := NewLFUEvictionStrategy[string](2)
		strategy.RecordInsertion("a", "b", "c", "d")

		keys := strategy.Evict()
		if len(keys) != 2 {
			t.Fatalf("expected 2 keys evicted, got %d: %v", len(keys), keys)
		}
	})

	t.Run("no eviction when maxSize is zero", func(t *testing.T) {
		t.Parallel()

		strategy := NewLFUEvictionStrategy[string](0)
		strategy.RecordInsertion("a", "b", "c")

		if keys := strategy.Evict(); len(keys) != 0 {
			t.Fatalf("expected no eviction with maxSize=0, got %v", keys)
		}
	})
}

func TestLFUEvictionStrategyRecordDeletion(t *testing.T) {
	t.Parallel()

	t.Run("deleted key is no longer eviction candidate", func(t *testing.T) {
		t.Parallel()

		strategy := NewLFUEvictionStrategy[string](1)
		strategy.RecordInsertion("a", "b")
		strategy.RecordDeletion("b")

		if keys := strategy.Evict(); len(keys) != 0 {
			t.Fatalf("expected no eviction after deletion brings count to maxSize, got %v", keys)
		}
	})

	t.Run("delete non-existent key is a no-op", func(t *testing.T) {
		t.Parallel()

		strategy := NewLFUEvictionStrategy[string](1)
		strategy.RecordInsertion("a", "b")
		strategy.RecordDeletion("missing")

		keys := strategy.Evict()
		if len(keys) != 1 {
			t.Fatalf("expected 1 eviction, got %d", len(keys))
		}
	})
}

func TestLFUEvictionStrategyClear(t *testing.T) {
	t.Parallel()

	strategy := NewLFUEvictionStrategy[string](10)
	strategy.RecordInsertion("a", "b", "c")
	strategy.RecordAccess("a", "b")
	strategy.Clear()

	if len(strategy.entries) != 0 {
		t.Fatalf("expected empty entries after Clear, got %d", len(strategy.entries))
	}
	if keys := strategy.Evict(); len(keys) != 0 {
		t.Fatalf("expected no evictions after Clear, got %v", keys)
	}
}

func TestLFUEvictionStrategyIsValid(t *testing.T) {
	t.Parallel()

	strategy := NewLFUEvictionStrategy[string](10)
	strategy.RecordInsertion("a")

	if !strategy.IsValid("a") {
		t.Fatal("expected IsValid to return true for any key")
	}
	if !strategy.IsValid("missing") {
		t.Fatal("expected IsValid to return true even for unknown keys")
	}
}

func TestLFUEvictionStrategyWithCache(t *testing.T) {
	t.Parallel()

	strategy := NewLFUEvictionStrategy[string](2)
	c := NewCache(WithEvictionStrategy[string, int](strategy))

	c.Set("a", 1)
	c.Set("b", 2)

	// Increase frequency for key "a" so key "b" is selected on overflow.
	if _, ok := c.Get("a"); !ok {
		t.Fatal("expected key a to exist")
	}
	if _, ok := c.Get("a"); !ok {
		t.Fatal("expected key a to exist")
	}

	c.Set("c", 3)

	if _, ok := c.Get("a"); !ok {
		t.Fatal("expected key a to be retained by LFU strategy")
	}
	if _, ok := c.Get("b"); ok {
		t.Fatal("expected key b to be evicted by LFU strategy")
	}
	if _, ok := c.Get("c"); !ok {
		t.Fatal("expected key c to exist")
	}
}
