package cache

import "testing"

func BenchmarkCacheSetOptions(b *testing.B) {
	copyBytes := func(in []byte) []byte {
		out := make([]byte, len(in))
		copy(out, in)
		return out
	}

	benchmarks := []struct {
		name string
		new  func() *Cache[int, []byte]
	}{
		{
			name: "default",
			new: func() *Cache[int, []byte] {
				return NewCache[int, []byte]()
			},
		},
		{
			name: "copy on set",
			new: func() *Cache[int, []byte] {
				return NewCache(WithCopyOnSet[int, []byte](copyBytes))
			},
		},
		{
			name: "max size with eviction",
			new: func() *Cache[int, []byte] {
				return NewCache(WithEvictionStrategy[int, []byte](NewLRUEvictionStrategy[int, []byte](4096)))
			},
		},
	}

	for _, bm := range benchmarks {
		bm := bm
		b.Run(bm.name, func(b *testing.B) {
			c := bm.new()
			value := []byte("benchmark-payload")
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				c.Set(i, value)
			}
		})
	}
}

func BenchmarkCacheGetOptions(b *testing.B) {
	copyBytes := func(in []byte) []byte {
		out := make([]byte, len(in))
		copy(out, in)
		return out
	}

	benchmarks := []struct {
		name string
		new  func() *Cache[int, []byte]
	}{
		{
			name: "default",
			new: func() *Cache[int, []byte] {
				return NewCache[int, []byte]()
			},
		},
		{
			name: "copy on get",
			new: func() *Cache[int, []byte] {
				return NewCache[int, []byte](WithCopyOnGet[int, []byte](copyBytes))
			},
		},
		{
			name: "max size with eviction",
			new: func() *Cache[int, []byte] {
				return NewCache[int, []byte](WithEvictionStrategy[int, []byte](NewLRUEvictionStrategy[int, []byte](4096)))
			},
		},
	}

	for _, bm := range benchmarks {
		bm := bm
		b.Run(bm.name, func(b *testing.B) {
			c := bm.new()
			value := []byte("benchmark-payload")
			for i := 0; i < 4096; i++ {
				c.Set(i, value)
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _ = c.Get(i % 4096)
			}
		})
	}
}

func BenchmarkCacheMixedOptions(b *testing.B) {
	copyBytes := func(in []byte) []byte {
		out := make([]byte, len(in))
		copy(out, in)
		return out
	}

	benchmarks := []struct {
		name string
		new  func() *Cache[int, []byte]
	}{
		{
			name: "default",
			new: func() *Cache[int, []byte] {
				return NewCache[int, []byte]()
			},
		},
		{
			name: "copy on set and get",
			new: func() *Cache[int, []byte] {
				return NewCache[int, []byte](
					WithCopyOnSet[int, []byte](copyBytes),
					WithCopyOnGet[int, []byte](copyBytes),
				)
			},
		},
		{
			name: "max size with eviction",
			new: func() *Cache[int, []byte] {
				return NewCache[int, []byte](WithEvictionStrategy[int, []byte](NewLRUEvictionStrategy[int, []byte](4096)))
			},
		},
	}

	for _, bm := range benchmarks {
		bm := bm
		b.Run(bm.name, func(b *testing.B) {
			c := bm.new()
			value := []byte("benchmark-payload")
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				key := i % 4096
				switch i % 3 {
				case 0:
					c.Set(key, value)
				case 1:
					_, _ = c.Get(key)
				case 2:
					c.CompareAndSwap(key, value, func(current, new []byte) bool {
						return len(current) <= len(new)
					})
				}
			}
		})
	}
}
