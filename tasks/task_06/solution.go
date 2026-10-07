package main

import "container/list"

type entry[K comparable, V any] struct {
	key   K
	value V
}
type LRUCache[K comparable, V any] struct {
	capacity int
	ll       list.List
	items    map[K]*list.Element
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{capacity: capacity, items: make(map[K]*list.Element)}
}
func (c *LRUCache[K, V]) Get(key K) (value V, ok bool) {
	if v, ok := c.items[key]; ok {
		el := v.Value.(entry[K, V])
		c.ll.MoveToFront(v)

		return el.value, true
	}

	return
}
func (c *LRUCache[K, V]) Set(key K, value V) {
	if c.capacity <= 0 {
		return
	}

	if v, ok := c.items[key]; ok {
		temp := v.Value.(entry[K, V])
		temp.value = value
		v.Value = temp
		return
	}

	if c.ll.Len() >= c.capacity {
		oldest := c.ll.Back()
		delete(c.items, oldest.Value.(entry[K, V]).key)
		c.ll.Remove(oldest)
	}

	newEntry := entry[K, V]{value: value, key: key}
	newElem := c.ll.PushFront(newEntry)
	c.items[key] = newElem
}

type LRU[K comparable, V any] interface {
	Get(K) (V, bool)
	Set(K, V)
}
