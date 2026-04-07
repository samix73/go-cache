package cache

// EvictionStrategy defines the interface for eviction strategies used in the cache.
type EvictionStrategy[K comparable] interface {
	// RecordAccess is called whenever multiple keys are accessed in the cache.
	RecordAccess(keys ...K)
	// RecordInsertion is called whenever multiple keys are inserted into the cache.
	RecordInsertion(keys ...K)
	// RecordDeletion is called whenever multiple keys are deleted from the cache.
	RecordDeletion(keys ...K)
	// Evict determines which key to evict from the cache based on the implemented strategy.
	Evict() []K
	// IsValid checks if the given key is valid according to the eviction strategy.
	IsValid(k K) bool
	// Clear resets the eviction strategy's internal state.
	Clear()
}
