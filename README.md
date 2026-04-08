[![Go Reference](https://pkg.go.dev/badge/github.com/samix73/go-cache.svg)](https://pkg.go.dev/github.com/samix73/go-cache)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/samix73/go-cache)
[![CI](https://github.com/samix73/go-cache/actions/workflows/ci.yml/badge.svg)](https://github.com/samix73/go-cache/actions/workflows/ci.yml)


# go-cache

A generic, thread-safe in-memory cache for Go with pluggable eviction strategies.

- Generic API: `Cache[K comparable, V any]`
- Concurrency-safe operations via `sync.RWMutex`
- Pluggable eviction strategies:
  - LRU (least recently used)
  - LFU (least frequently used)
  - TTL (time-to-live)
  - Random
  - Composite (combine multiple strategies)
- Optional copy hooks for safer mutable values (`WithCopyOnSet`, `WithCopyOnGet`)

## Installation

```bash
go get github.com/samix73/go-cache
```

## Quick Start

```go
package main

import (
  "fmt"

  cache "github.com/samix73/go-cache"
)

func main() {
  c := cache.NewCache[string, int]()
  c.Set("apples", 3)

  v, ok := c.Get("apples")
  fmt.Println(v, ok) // 3 true
}
```

## Core API

```go
Set(key, value)
MSet(map[K]V)
Get(key) (V, bool)
MGet(keys ...K) map[K]V
Delete(key)
Clear()
CompareAndSwap(key, newValue, compareFn) bool
StartEvictionRoutine(ctx, interval) error
```

## API Overview

`NewCache(opts ...CacheOptions[K, V])`
- Creates a new cache instance.

`Set(key, value)`
- Inserts or replaces a single key.

`MSet(pairs)`
- Inserts or replaces multiple keys in one call.

`Get(key) (V, bool)`
- Returns `(value, true)` when present and valid.
- Returns `(zeroValue, false)` for missing or invalid entries.

`MGet(keys ...K) map[K]V`
- Returns only keys that exist and are valid.

`Delete(key)`
- Removes a key from cache and strategy state.

`Clear()`
- Clears cache storage and resets strategy state.

`CompareAndSwap(key, value, compareFn) bool`
- Atomically updates a key when `compareFn(current, new)` returns true.

`StartEvictionRoutine(ctx, interval) error`
- Starts periodic eviction until `ctx` is canceled.
- Returns an error if no eviction strategy is configured.

## Read-Through Cache

`ReadThroughCache[K, V]` wraps a `Cache` and a `Loader` to provide read-through
semantics: on a cache miss the loader is called to fetch the value from a backing
store (database, HTTP API, file, etc.), the result is stored in the cache, and
then returned to the caller.

### Thundering-herd protection

Concurrent misses for the same key result in **only one loader call**.  All
waiting goroutines receive the same value once it arrives (backed by
[`singleflight`](https://pkg.go.dev/golang.org/x/sync/singleflight)).

### Loader interface

```go
type Loader[K comparable, V any] interface {
    Load(ctx context.Context, key K) (V, bool, error)
}
```

Return `(value, true, nil)` when found, `(zero, false, nil)` when not found, or
`(zero, false, err)` on error.  A `LoaderFunc` adapter is provided for simple
function-based loaders.

### Quick example

```go
package main

import (
    "context"
    "fmt"

    cache "github.com/samix73/go-cache"
)

// dbLoader simulates loading from a database.
type dbLoader struct {
    db map[string]int
}

func (d *dbLoader) Load(_ context.Context, key string) (int, bool, error) {
    v, ok := d.db[key]
    return v, ok, nil
}

func main() {
    c := cache.NewCache[string, int]()
    loader := &dbLoader{db: map[string]int{"apples": 3}}

    rtc := cache.NewReadThroughCache(c, loader)

    v, ok, err := rtc.Get(context.Background(), "apples")
    fmt.Println(v, ok, err) // 3 true <nil>

    // Second call hits the in-memory cache; loader is not called again.
    v, ok, err = rtc.Get(context.Background(), "apples")
    fmt.Println(v, ok, err) // 3 true <nil>
}
```

You can also use the `LoaderFunc` adapter instead of defining a type:

```go
loader := cache.LoaderFunc[string, int](func(ctx context.Context, key string) (int, bool, error) {
    // fetch from backing store …
    return 42, true, nil
})
rtc := cache.NewReadThroughCache(c, loader)
```

### API

`NewReadThroughCache(c *Cache[K,V], loader Loader[K,V]) *ReadThroughCache[K,V]`

`Get(ctx context.Context, key K) (V, bool, error)`
- Returns `(value, true, nil)` when found in cache or loaded successfully.
- Returns `(zero, false, nil)` when the loader reports not-found (result is not stored).
- Returns `(zero, false, err)` when the loader errors (result is not stored).

`Set(key K, value V)` — pass-through to the underlying cache.

`Delete(key K)` — pass-through to the underlying cache.

## Eviction Strategies

- `NewLRUEvictionStrategy(maxSize)`: evicts least-recently-used keys when above capacity.
- `NewLFUEvictionStrategy(maxSize)`: evicts least-frequently-used keys when above capacity.
- `NewTTLEvictionStrategy(ttl)`: treats entries as invalid after TTL from insertion.
- `NewRandomEvictionStrategy(maxSize)`: evicts random keys when above capacity.
- `NewCompositeEvictionStrategy(strategies...)`: combines multiple strategies.

## Options

- `WithEvictionStrategy(strategy)`: plugs in eviction behavior.
- `WithCopyOnSet(copyFn)`: copies values before storing.
- `WithCopyOnGet(copyFn)`: copies values before returning.
- `WithDisableEvictionOnSet()`: disables eviction during `Set`/`MSet` only.

## Development

Tests stay in the repository root as `*_test.go` files.

Run tests:

```bash
go test ./...
```

Run race check:

```bash
go test -race -failfast ./...
```

Run benchmarks:

```bash
go test -bench . -v
```

## License

MIT. See [LICENSE](LICENSE).
