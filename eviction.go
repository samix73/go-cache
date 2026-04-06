package cache

// EvictionStrategy defines the interface for eviction strategies used in the cache.
type EvictionStrategy[K comparable, V any] interface {
	// RecordAccess is called whenever a key is accessed in the cache.
	RecordAccess(key K)
	// RecordInsertion is called whenever a key is inserted into the cache.
	RecordInsertion(key K)
	// RecordDeletion is called whenever a key is deleted from the cache.
	RecordDeletion(key K)
	// Evict determines which key to evict from the cache based on the implemented strategy.
	// It returns the key to evict and a boolean indicating whether eviction should occur.
	Evict() (key K, shouldEvict bool)
	// Clear resets the eviction strategy's internal state.
	Clear()
}
