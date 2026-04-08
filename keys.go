package cache

import "fmt"

// singleflightKey converts any comparable key to a string suitable for use
// with singleflight.Group. fmt.Sprintf("%v", k) is deterministic for all
// built-in comparable types and produces unique strings for distinct values.
func singleflightKey[K comparable](k K) string {
	return fmt.Sprintf("%v", k)
}
