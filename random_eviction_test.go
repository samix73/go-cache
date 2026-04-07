package cache

import (
	"testing"
)

func TestRandomEvictionStrategyEvict(t *testing.T) {
	t.Parallel()

	t.Run("does not evict when under max size", func(t *testing.T) {
		t.Parallel()

		strategy := NewRandomEvictionStrategy[string, int](3)
		strategy.RecordInsertion("a", "b")

		if keys := strategy.Evict(); len(keys) != 0 {
			t.Fatalf("expected no eviction, got %v", keys)
		}
	})

	t.Run("does not evict when at max size", func(t *testing.T) {
		t.Parallel()

		strategy := NewRandomEvictionStrategy[string, int](3)
		strategy.RecordInsertion("a", "b", "c")

		if keys := strategy.Evict(); len(keys) != 0 {
			t.Fatalf("expected no eviction at exact max size, got %v", keys)
		}
	})

	t.Run("evicts correct number of keys when over max size", func(t *testing.T) {
		t.Parallel()

		strategy := NewRandomEvictionStrategy[string, int](2)
		strategy.RecordInsertion("a", "b", "c", "d")

		keys := strategy.Evict()
		if len(keys) != 2 {
			t.Fatalf("expected 2 keys evicted, got %d: %v", len(keys), keys)
		}
	})

	t.Run("evicted keys are distinct", func(t *testing.T) {
		t.Parallel()

		strategy := NewRandomEvictionStrategy[string, int](1)
		strategy.RecordInsertion("a", "b", "c", "d", "e")

		keys := strategy.Evict()
		if len(keys) != 4 {
			t.Fatalf("expected 4 evictions, got %d", len(keys))
		}

		seen := make(map[string]struct{}, len(keys))
		for _, k := range keys {
			if _, dup := seen[k]; dup {
				t.Fatalf("duplicate key in eviction result: %q", k)
			}
			seen[k] = struct{}{}
		}
	})

	t.Run("evicted keys exist in the strategy", func(t *testing.T) {
		t.Parallel()

		strategy := NewRandomEvictionStrategy[string, int](1)
		strategy.RecordInsertion("a", "b", "c")

		keys := strategy.Evict()
		known := map[string]struct{}{"a": {}, "b": {}, "c": {}}
		for _, k := range keys {
			if _, exists := known[k]; !exists {
				t.Fatalf("evicted unknown key %q", k)
			}
		}
	})

	t.Run("no eviction when maxSize is zero", func(t *testing.T) {
		t.Parallel()

		strategy := NewRandomEvictionStrategy[string, int](0)
		strategy.RecordInsertion("a", "b", "c")

		if keys := strategy.Evict(); len(keys) != 0 {
			t.Fatalf("expected no eviction with maxSize=0, got %v", keys)
		}
	})
}

func TestRandomEvictionStrategyRecordDeletion(t *testing.T) {
	t.Parallel()

	t.Run("deleted key is no longer eviction candidate", func(t *testing.T) {
		t.Parallel()

		strategy := NewRandomEvictionStrategy[string, int](1)
		strategy.RecordInsertion("a", "b")
		strategy.RecordDeletion("b")

		// Only "a" and "b" were inserted; after deleting "b", only "a" remains.
		// With maxSize=1 and 1 key, nothing should be evicted.
		if keys := strategy.Evict(); len(keys) != 0 {
			t.Fatalf("expected no eviction after deletion brings count to maxSize, got %v", keys)
		}
	})

	t.Run("delete non-existent key is a no-op", func(t *testing.T) {
		t.Parallel()

		strategy := NewRandomEvictionStrategy[string, int](1)
		strategy.RecordInsertion("a", "b")
		strategy.RecordDeletion("missing")

		keys := strategy.Evict()
		if len(keys) != 1 {
			t.Fatalf("expected 1 eviction, got %d", len(keys))
		}
	})

	t.Run("internal index is consistent after deletion", func(t *testing.T) {
		t.Parallel()

		strategy := NewRandomEvictionStrategy[string, int](10)
		strategy.RecordInsertion("a", "b", "c")
		strategy.RecordDeletion("b")

		if len(strategy.keys) != 2 {
			t.Fatalf("expected 2 keys after deletion, got %d", len(strategy.keys))
		}
		for i, k := range strategy.keys {
			if strategy.index[k] != i {
				t.Fatalf("index mismatch: key %q has index %d, expected %d", k, strategy.index[k], i)
			}
		}
	})
}

func TestRandomEvictionStrategyClear(t *testing.T) {
	t.Parallel()

	strategy := NewRandomEvictionStrategy[string, int](10)
	strategy.RecordInsertion("a", "b", "c")
	strategy.Clear()

	if len(strategy.keys) != 0 {
		t.Fatalf("expected empty keys after Clear, got %d", len(strategy.keys))
	}
	if len(strategy.index) != 0 {
		t.Fatalf("expected empty index after Clear, got %d", len(strategy.index))
	}
	if keys := strategy.Evict(); len(keys) != 0 {
		t.Fatalf("expected no evictions after Clear, got %v", keys)
	}
}

func TestRandomEvictionStrategyIsValid(t *testing.T) {
	t.Parallel()

	strategy := NewRandomEvictionStrategy[string, int](10)
	strategy.RecordInsertion("a")

	if !strategy.IsValid("a") {
		t.Fatal("expected IsValid to return true for any key")
	}
	if !strategy.IsValid("missing") {
		t.Fatal("expected IsValid to return true even for unknown keys")
	}
}

func TestRandomEvictionStrategyWithCache(t *testing.T) {
	t.Parallel()

	t.Run("cache respects max size with random eviction", func(t *testing.T) {
		t.Parallel()

		const maxSize = 3
		strategy := NewRandomEvictionStrategy[string, int](maxSize)
		c := NewCache(WithEvictionStrategy[string, int](strategy))

		for i, key := range []string{"a", "b", "c", "d", "e"} {
			c.Set(key, i)
		}

		if size := cacheSize(c); size != maxSize {
			t.Fatalf("expected cache size %d, got %d", maxSize, size)
		}
	})

	t.Run("duplicate insertion does not grow internal state", func(t *testing.T) {
		t.Parallel()

		strategy := NewRandomEvictionStrategy[string, int](10)
		strategy.RecordInsertion("a")
		strategy.RecordInsertion("a")

		if len(strategy.keys) != 1 {
			t.Fatalf("expected 1 key after duplicate insertion, got %d", len(strategy.keys))
		}
	})
}
