package cache

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestCacheGetSet(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		opts  []CacheOptions[string, int]
		key   string
		value int
	}{
		{
			name:  "default options",
			key:   "a",
			value: 10,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := NewCache(tc.opts...)
			c.Set(tc.key, tc.value)

			got, ok := c.Get(tc.key)
			if !ok {
				t.Fatalf("expected key %q to exist", tc.key)
			}
			if got != tc.value {
				t.Fatalf("expected value %d, got %d", tc.value, got)
			}
		})
	}
}

func TestCacheMissingGet(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		key       string
		zeroValue int
	}{
		{name: "missing key", key: "missing", zeroValue: 0},
		{name: "another missing key", key: "none", zeroValue: 0},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := NewCache[string, int]()
			got, ok := c.Get(tc.key)
			if ok {
				t.Fatalf("expected key %q to be missing", tc.key)
			}
			if got != tc.zeroValue {
				t.Fatalf("expected zero value %d, got %d", tc.zeroValue, got)
			}
		})
	}
}

func TestCacheDeleteAndClear(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "delete existing key",
			run: func(t *testing.T) {
				c := NewCache[string, int]()
				c.Set("a", 1)
				c.Delete("a")

				_, ok := c.Get("a")
				if ok {
					t.Fatal("expected key to be deleted")
				}
			},
		},
		{
			name: "clear removes all keys and resets strategy",
			run: func(t *testing.T) {
				strategy := NewLRUEvictionStrategy[string, int](10)
				c := NewCache(WithEvictionStrategy[string, int](strategy))
				c.Set("a", 1)
				c.Set("b", 2)

				c.Clear()

				if size := cacheSize(c); size != 0 {
					t.Fatalf("expected cache size 0 after Clear, got %d", size)
				}
				if len(strategy.lookup) != 0 {
					t.Fatalf("expected strategy keys to be cleared, got %d", len(strategy.lookup))
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.run(t)
		})
	}
}

func TestCacheCompareAndSwap(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		seed         *int
		newValue     int
		compareFn    func(current, new int) bool
		wantSwapped  bool
		wantValue    int
		wantValueSet bool
	}{
		{
			name:         "nil compare function",
			seed:         new(1),
			newValue:     5,
			compareFn:    nil,
			wantSwapped:  false,
			wantValue:    1,
			wantValueSet: true,
		},
		{
			name:         "missing key",
			seed:         nil,
			newValue:     5,
			compareFn:    func(current, new int) bool { return true },
			wantSwapped:  false,
			wantValueSet: false,
		},
		{
			name:         "compare returns false",
			seed:         new(3),
			newValue:     7,
			compareFn:    func(current, new int) bool { return false },
			wantSwapped:  false,
			wantValue:    3,
			wantValueSet: true,
		},
		{
			name:         "compare returns true",
			seed:         new(3),
			newValue:     7,
			compareFn:    func(current, new int) bool { return current < new },
			wantSwapped:  true,
			wantValue:    7,
			wantValueSet: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := NewCache[string, int]()
			if tc.seed != nil {
				c.Set("k", *tc.seed)
			}

			swapped, err := c.CompareAndSwap("k", tc.newValue, tc.compareFn)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if swapped != tc.wantSwapped {
				t.Fatalf("expected swapped=%v, got %v", tc.wantSwapped, swapped)
			}

			got, ok := c.Get("k")
			if ok != tc.wantValueSet {
				t.Fatalf("expected exists=%v, got %v", tc.wantValueSet, ok)
			}
			if tc.wantValueSet && got != tc.wantValue {
				t.Fatalf("expected value %d, got %d", tc.wantValue, got)
			}
		})
	}
}

func TestCacheCopyOptions(t *testing.T) {
	t.Parallel()

	copySlice := func(in []int) []int {
		out := make([]int, len(in))
		copy(out, in)
		return out
	}

	testCases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "copy on set isolates stored value",
			run: func(t *testing.T) {
				c := NewCache(WithCopyOnSet[int](copySlice))
				input := []int{1, 2, 3}
				c.Set(1, input)
				input[0] = 99

				got, ok := c.Get(1)
				if !ok {
					t.Fatal("expected key to exist")
				}
				if got[0] != 1 {
					t.Fatalf("expected stored value to remain unchanged, got %v", got)
				}
			},
		},
		{
			name: "copy on get isolates returned value",
			run: func(t *testing.T) {
				c := NewCache[int, []int](WithCopyOnGet[int, []int](copySlice))
				c.Set(1, []int{1, 2, 3})

				got, ok := c.Get(1)
				if !ok {
					t.Fatal("expected key to exist")
				}
				got[0] = 99

				again, ok := c.Get(1)
				if !ok {
					t.Fatal("expected key to exist on second read")
				}
				if again[0] != 1 {
					t.Fatalf("expected cached value to remain unchanged, got %v", again)
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.run(t)
		})
	}
}

func TestCacheMaxSizeEviction(t *testing.T) {
	t.Parallel()

	strategy := NewLRUEvictionStrategy[string, int](2)
	c := NewCache(WithEvictionStrategy[string, int](strategy))

	c.Set("a", 1)
	c.Set("b", 2)
	// Access "a" so "b" becomes the least recently used key.
	if _, ok := c.Get("a"); !ok {
		t.Fatal("expected key a to exist before eviction")
	}
	c.Set("c", 3)

	if size := cacheSize(c); size != 2 {
		t.Fatalf("expected cache size 2, got %d", size)
	}

	if _, ok := c.Get("b"); ok {
		t.Fatal("expected least recently used key to be evicted")
	}
	if _, ok := c.Get("a"); !ok {
		t.Fatal("expected key a to remain")
	}
	if _, ok := c.Get("c"); !ok {
		t.Fatal("expected key c to remain")
	}
}

func TestCacheDisableEvictionOnSet(t *testing.T) {
	t.Parallel()

	// With DisableEvictionOnSet, inserting beyond the max size should not evict entries.
	strategy := NewLRUEvictionStrategy[string, int](2)
	c := NewCache(
		WithEvictionStrategy[string, int](strategy),
		WithDisableEvictionOnSet[string, int](),
	)

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3) // would normally evict one entry, but should not here

	if size := cacheSize(c); size != 3 {
		t.Fatalf("expected cache size 3 (eviction disabled on set), got %d", size)
	}

	for _, key := range []string{"a", "b", "c"} {
		if _, ok := c.Get(key); !ok {
			t.Fatalf("expected key %q to remain when eviction on set is disabled", key)
		}
	}
}

func TestCacheDisableEvictionOnSetBackgroundRoutineStillEvicts(t *testing.T) {
	t.Parallel()

	// Even with DisableEvictionOnSet, the background eviction routine should still evict entries.
	strategy := NewLRUEvictionStrategy[string, int](2)
	c := NewCache(
		WithEvictionStrategy[string, int](strategy),
		WithDisableEvictionOnSet[string, int](),
	)

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3) // goes in without eviction because the option is set

	if size := cacheSize(c); size != 3 {
		t.Fatalf("expected cache size 3 before background eviction, got %d", size)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		_ = c.StartEvictionRoutine(ctx, 10*time.Millisecond)
	}()

	// Wait until the cache shrinks to the configured max size (2).
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cacheSize(c) <= 2 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if size := cacheSize(c); size > 2 {
		t.Fatalf("expected background routine to evict down to max size 2, got %d", size)
	}
}

func TestCacheConcurrentAccess(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		opts []CacheOptions[int, int]
		max  int
	}{
		{
			name: "default",
		},
		{
			name: "bounded with lru eviction",
			opts: []CacheOptions[int, int]{
				WithEvictionStrategy[int, int](NewLRUEvictionStrategy[int, int](128)),
			},
			max: 128,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := NewCache(tc.opts...)
			const workers = 24
			const opsPerWorker = 2000

			var wg sync.WaitGroup
			start := make(chan struct{})

			for worker := range workers {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-start

					for op := 0; op < opsPerWorker; op++ {
						key := (worker*opsPerWorker + op) % 256
						switch op % 4 {
						case 0:
							c.Set(key, op)
						case 1:
							_, _ = c.Get(key)
						case 2:
							_, _ = c.CompareAndSwap(key, op, func(current, new int) bool {
								return current <= new
							})
						case 3:
							c.Delete(key)
						}
					}
				}()
			}

			close(start)
			wg.Wait()

			if tc.max > 0 {
				if size := cacheSize(c); size > tc.max {
					t.Fatalf("expected cache size <= %d, got %d", tc.max, size)
				}
			}
		})
	}
}

func cacheSize[K comparable, V any](c *Cache[K, V]) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.storage)
}
