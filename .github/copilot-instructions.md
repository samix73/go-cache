# Project Guidelines

## Build And Test
- Go version: 1.26 (see go.mod).
- Run unit tests: go test ./...
- Preferred full check before merging: go test -race -failfast ./...
- Run benchmarks: go test -bench . -v

## Architecture
- This repository is a single Go package implementing a generic in-memory cache: Cache[K comparable, V any].
- Core cache API and locking live in cache.go.
- Eviction behavior is pluggable via EvictionStrategy in eviction.go.
- Implemented strategies:
  - lru_eviction.go: size-bounded least recently used eviction
  - ttl_eviction.go: time-to-live expiration
  - random_eviction.go: size-bounded random eviction
  - composite_eviction.go: combines multiple strategies and deduplicates keys to evict
- Configuration uses functional options in options.go.
- Backend interfaces (Loader, etc.) live in backend.go; they decouple cache policy wrappers from concrete backing stores.
- Policy wrappers sit on top of Cache[K,V] and add load/store semantics without touching the eviction layer:
  - read_through.go: ReadThroughCache[K,V] – reads from loader on miss, stores result, singleflight thundering-herd guard.
- To add a new policy (e.g. write-through, write-back):
  1. Add the required backend interface(s) to backend.go (e.g. Writer[K,V]).
  2. Create a new file (e.g. write_through.go) with the wrapper struct and functional options.
  3. Add a corresponding _test.go following the table-driven parallel test style.
  4. Update README.md and this file.

## Conventions
- Keep cache operations thread-safe: cache uses sync.RWMutex, and stateful eviction strategies must manage their own locking.
- Strategy hooks (RecordInsertion/RecordAccess/RecordDeletion/Evict/IsValid) should remain fast and non-blocking because they run on cache hot paths.
- Preserve generic APIs and zero-value behavior for missing keys.
- Keep compile-time interface assertions for strategy implementations (var _ EvictionStrategy[...] = (*Type[...])(nil)).
- Prefer table-driven tests with parallel subtests, consistent with existing tests.

## Pitfalls And Gotchas
- WithDisableEvictionOnSet disables eviction during Set/MSet only. Background eviction via StartEvictionRoutine still evicts.
- Composite strategy must ignore nil child strategies and deduplicate eviction keys.
- Map iteration order is non-deterministic; tests should assert counts/invariants rather than specific key order when maps are involved.

## Key Files
- cache.go: primary cache behavior and synchronization model
- options.go: functional options and behavior toggles
- lru_eviction.go: reference for size-based stateful strategy implementation
- cache_test.go: test style and concurrency patterns
- backend.go: Loader (and future Writer/Deleter) backend interfaces
- read_through.go: ReadThroughCache wrapper with singleflight thundering-herd guard
- read_through_test.go: tests for ReadThroughCache
