package cache

import (
	"context"
	"testing"
	"time"
)

func TestTTLEvictionStrategyIsValid(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		ttl       time.Duration
		sleep     time.Duration
		wantValid bool
	}{
		{
			name:      "entry is valid before TTL expires",
			ttl:       time.Hour,
			sleep:     0,
			wantValid: true,
		},
		{
			name:      "entry is invalid after TTL expires",
			ttl:       10 * time.Millisecond,
			sleep:     50 * time.Millisecond,
			wantValid: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			strategy := NewTTLEvictionStrategy[string](tc.ttl)
			strategy.RecordInsertion("key")

			if tc.sleep > 0 {
				time.Sleep(tc.sleep)
			}

			if got := strategy.IsValid("key"); got != tc.wantValid {
				t.Fatalf("IsValid() = %v, want %v", got, tc.wantValid)
			}
		})
	}
}

func TestTTLEvictionStrategyIsValidMissingKey(t *testing.T) {
	t.Parallel()

	strategy := NewTTLEvictionStrategy[string](time.Hour)
	if strategy.IsValid("missing") {
		t.Fatal("expected missing key to be invalid")
	}
}

func TestTTLEvictionStrategyEvict(t *testing.T) {
	t.Parallel()

	strategy := NewTTLEvictionStrategy[string](20 * time.Millisecond)
	strategy.RecordInsertion("a", "b")

	// No keys should be expired yet.
	if keys := strategy.Evict(); len(keys) != 0 {
		t.Fatalf("expected no expired keys, got %v", keys)
	}

	time.Sleep(50 * time.Millisecond)

	expired := strategy.Evict()
	if len(expired) != 2 {
		t.Fatalf("expected 2 expired keys, got %d: %v", len(expired), expired)
	}
}

func TestTTLEvictionStrategyRecordDeletion(t *testing.T) {
	t.Parallel()

	strategy := NewTTLEvictionStrategy[string](time.Hour)
	strategy.RecordInsertion("a")
	strategy.RecordDeletion("a")

	if strategy.IsValid("a") {
		t.Fatal("expected deleted key to be invalid")
	}
}

func TestTTLEvictionStrategyClear(t *testing.T) {
	t.Parallel()

	strategy := NewTTLEvictionStrategy[string](time.Hour)
	strategy.RecordInsertion("a", "b")
	strategy.Clear()

	if strategy.IsValid("a") {
		t.Fatal("expected cleared key to be invalid")
	}
	if len(strategy.insertedAt) != 0 {
		t.Fatalf("expected empty insertedAt after Clear, got %d entries", len(strategy.insertedAt))
	}
}

func TestTTLEvictionStrategyWithCache(t *testing.T) {
	t.Parallel()

	strategy := NewTTLEvictionStrategy[string](30 * time.Millisecond)
	c := NewCache(WithEvictionStrategy[string, int](strategy))

	c.Set("a", 1)
	c.Set("b", 2)

	// Both keys should be readable before TTL expires.
	if _, ok := c.Get("a"); !ok {
		t.Fatal("expected key a to exist before TTL")
	}
	if _, ok := c.Get("b"); !ok {
		t.Fatal("expected key b to exist before TTL")
	}

	time.Sleep(60 * time.Millisecond)

	// After TTL expires, keys should not be readable.
	if _, ok := c.Get("a"); ok {
		t.Fatal("expected key a to be expired")
	}
	if _, ok := c.Get("b"); ok {
		t.Fatal("expected key b to be expired")
	}
}

func TestTTLEvictionStrategyEvictRemovesFromCache(t *testing.T) {
	t.Parallel()

	strategy := NewTTLEvictionStrategy[string](20 * time.Millisecond)
	c := NewCache(
		WithEvictionStrategy[string, int](strategy),
		WithDisableEvictionOnSet[string, int](),
	)

	c.Set("a", 1)
	c.Set("b", 2)

	time.Sleep(50 * time.Millisecond)

	// Run the eviction routine briefly to flush expired entries.
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = c.StartEvictionRoutine(ctx, 10*time.Millisecond)
	}()
	<-done

	if size := cacheSize(c); size != 0 {
		t.Fatalf("expected cache to be empty after TTL eviction, got size %d", size)
	}
}
