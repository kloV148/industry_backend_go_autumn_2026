package main

import (
	"container/list"
	"sync"
)

type entry[K comparable, V any] struct {
	key   K
	value V
}

type LRUCache[K comparable, V any] struct {
	capacity int
	ll       list.List
	items    map[K]*list.Element
	m        sync.Mutex
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		capacity: capacity,
		items:    map[K]*list.Element{},
	}
}

func (c *LRUCache[K, V]) Get(key K) (value V, ok bool) {
	c.m.Lock()
	defer c.m.Unlock()

	element, ok := c.items[key]

	if ok {
		c.ll.MoveToFront(element)

		item := element.Value.(*entry[K, V])

		return item.value, ok
	}

	return value, false
}

func (c *LRUCache[K, V]) Set(key K, value V) {
	if c.capacity <= 0 {
		return
	}

	c.m.Lock()
	defer c.m.Unlock()
	
	_, keyExist := c.items[key]

	if keyExist {
		item := c.items[key].Value.(*entry[K, V])

		item.value = value
	} else {
		if c.capacity <= c.ll.Len() {
			element := c.ll.Back()
			item := element.Value.(*entry[K, V])

			delete(c.items, item.key)
			c.ll.Remove(element)
		}

		item := &entry[K, V]{value: value, key: key}
		element := c.ll.PushFront(item)

		c.items[key] = element
	}
}

type LRU[K comparable, V any] interface {
	Get(K) (V, bool)
	Set(K, V)
}
