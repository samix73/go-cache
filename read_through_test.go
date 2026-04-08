package cache

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// mockLoader is a test helper that records calls and returns configured results.
type mockLoader[K comparable, V any] struct {
	mu      sync.Mutex
	calls   int
	value   V
	found   bool
	loadErr error
	// loadFn overrides value/found/loadErr if set.
	loadFn func(ctx context.Context, key K) (V, bool, error)
}

func (m *mockLoader[K, V]) Load(ctx context.Context, key K) (V, bool, error) {
	m.mu.Lock()
	m.calls++
	m.mu.Unlock()

	if m.loadFn != nil {
		return m.loadFn(ctx, key)
	}
	return m.value, m.found, m.loadErr
}

func (m *mockLoader[K, V]) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}

func TestReadThroughCacheHit(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		key   string
		value int
	}{
		{name: "simple hit", key: "a", value: 1},
		{name: "zero value hit", key: "b", value: 0},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			loader := &mockLoader[string, int]{found: true, value: 99}
			c := NewCache[string, int]()
			c.Set(tc.key, tc.value)
			rtc := NewReadThroughCache(c, loader)

			got, ok, err := rtc.Get(t.Context(), tc.key)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !ok {
				t.Fatalf("expected key %q to be found", tc.key)
			}
			if got != tc.value {
				t.Fatalf("expected value %v, got %v", tc.value, got)
			}
			if loader.callCount() != 0 {
				t.Fatalf("expected loader not to be called on cache hit, got %d calls", loader.callCount())
			}
		})
	}
}

func TestReadThroughCacheMissCallsLoaderAndStores(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		key         string
		loaderValue int
	}{
		{name: "loader returns value", key: "x", loaderValue: 42},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			loader := &mockLoader[string, int]{value: tc.loaderValue, found: true}
			c := NewCache[string, int]()
			rtc := NewReadThroughCache(c, loader)

			got, ok, err := rtc.Get(t.Context(), tc.key)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !ok {
				t.Fatalf("expected found=true from loader")
			}
			if got != tc.loaderValue {
				t.Fatalf("expected loader value %d, got %d", tc.loaderValue, got)
			}
			if loader.callCount() != 1 {
				t.Fatalf("expected loader to be called once, got %d", loader.callCount())
			}

			// Second Get must hit the cache; loader must not be called again.
			got2, ok2, err2 := rtc.Get(t.Context(), tc.key)
			if err2 != nil {
				t.Fatalf("unexpected error on second Get: %v", err2)
			}
			if !ok2 || got2 != tc.loaderValue {
				t.Fatalf("expected cached value %d, got %d (ok=%v)", tc.loaderValue, got2, ok2)
			}
			if loader.callCount() != 1 {
				t.Fatalf("expected loader to still be called once, got %d", loader.callCount())
			}
		})
	}
}

func TestReadThroughCacheNotFoundDoesNotStore(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		key         string
		loaderValue int // value the loader returns alongside found=false
	}{
		{name: "loader reports not found, zero value", key: "missing", loaderValue: 0},
		{name: "loader reports not found, non-zero value", key: "missing2", loaderValue: 99},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			loader := &mockLoader[string, int]{found: false, value: tc.loaderValue}
			c := NewCache[string, int]()
			rtc := NewReadThroughCache(c, loader)

			got, ok, err := rtc.Get(t.Context(), tc.key)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ok {
				t.Fatal("expected found=false when loader reports not found")
			}
			if got != 0 {
				t.Fatalf("expected zero value, got %d", got)
			}

			// Key must not have been stored in the underlying cache.
			if _, inCache := c.Get(tc.key); inCache {
				t.Fatal("not-found result must not be stored in cache")
			}
		})
	}
}

func TestReadThroughCacheLoaderErrorDoesNotStore(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		key     string
		loadErr error
	}{
		{name: "loader returns error", key: "boom", loadErr: errors.New("backend error")},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			loader := &mockLoader[string, int]{loadErr: tc.loadErr}
			c := NewCache[string, int]()
			rtc := NewReadThroughCache(c, loader)

			got, ok, err := rtc.Get(t.Context(), tc.key)
			if err == nil {
				t.Fatal("expected error to be propagated")
			}
			if !errors.Is(err, tc.loadErr) {
				t.Fatalf("expected error %v, got %v", tc.loadErr, err)
			}
			if ok {
				t.Fatal("expected found=false on loader error")
			}
			if got != 0 {
				t.Fatalf("expected zero value on error, got %d", got)
			}

			// Key must not have been stored.
			if _, inCache := c.Get(tc.key); inCache {
				t.Fatal("error result must not be stored in cache")
			}
		})
	}
}

func TestReadThroughCacheConcurrentMissCallsLoaderOnce(t *testing.T) {
	t.Parallel()

	const goroutines = 50
	const key = "shared-key"
	const loaderValue = 7

	var loadCallCount atomic.Int64
	// loaderStarted is closed by the loader once it begins executing, so the test
	// can confirm the singleflight call is in-flight before starting other goroutines.
	loaderStarted := make(chan struct{})
	// release is closed by the test to allow the loader to return.
	release := make(chan struct{})

	var startOnce sync.Once
	loader := LoaderFunc[string, int](func(ctx context.Context, k string) (int, bool, error) {
		loadCallCount.Add(1)
		startOnce.Do(func() { close(loaderStarted) })
		<-release
		return loaderValue, true, nil
	})

	c := NewCache[string, int]()
	rtc := NewReadThroughCache(c, loader)

	results := make([]int, goroutines)
	founds := make([]bool, goroutines)
	errs := make([]error, goroutines)

	var wg sync.WaitGroup
	// Goroutine 0 fires first to guarantee the singleflight call is in-flight
	// before the remaining goroutines start.
	wg.Add(1)
	go func() {
		defer wg.Done()
		results[0], founds[0], errs[0] = rtc.Get(t.Context(), key)
	}()

	// Wait until the loader is actually executing (singleflight call is in-flight).
	<-loaderStarted

	// Now start the remaining goroutines; they will all queue in the same
	// in-flight singleflight call and must not trigger an additional load.
	for i := 1; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], founds[idx], errs[idx] = rtc.Get(t.Context(), key)
		}(i)
	}

	// Give goroutines time to register in singleflight before releasing the loader.
	// A small sleep is more reliable than Gosched here because all goroutines need
	// to advance past their cache-miss check and into DoChan before we unblock the loader.
	time.Sleep(10 * time.Millisecond)
	close(release)
	wg.Wait()

	for i := range goroutines {
		if errs[i] != nil {
			t.Errorf("goroutine %d: unexpected error: %v", i, errs[i])
		}
		if !founds[i] {
			t.Errorf("goroutine %d: expected found=true", i)
		}
		if results[i] != loaderValue {
			t.Errorf("goroutine %d: expected value %d, got %d", i, loaderValue, results[i])
		}
	}

	// singleflight must have collapsed concurrent misses into a single load call.
	if n := loadCallCount.Load(); n != 1 {
		t.Fatalf("expected loader to be called exactly once, got %d calls", n)
	}
	// Value must be stored in the underlying cache after the load.
	if _, ok := c.Get(key); !ok {
		t.Fatal("expected value to be stored in cache after concurrent miss")
	}
}
