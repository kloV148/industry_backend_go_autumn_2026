package main

type Cache[K comparable, V any] struct {
	capacity int
	items    map[K]V
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	cache := Cache[K, V]{
		capacity: capacity,
		items:    map[K]V{},
	}
	return &cache
}

func (c *Cache[K, V]) Get(k K) (v V, ok bool) {
	v, ok = c.items[k]

	return v, ok
}

func (c *Cache[K, V]) Set(k K, v V) bool {
	capacity := c.capacity

	if capacity <= 0 {
		return false
	}

	_, keyExist := c.items[k]

	if !keyExist && capacity <= len(c.items) {
		return false
	}

	c.items[k] = v
	return true
}
