[![Go Reference](https://pkg.go.dev/badge/github.com/samix73/go-cache.svg)](https://pkg.go.dev/github.com/samix73/go-cache)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/samix73/go-cache)


# go-cache

A generic, thread-safe in-memory cache for Go with pluggable eviction strategies.

GoDoc: [pkg.go.dev/github.com/samix73/go-cache](https://pkg.go.dev/github.com/samix73/go-cache)

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
