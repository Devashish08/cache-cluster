// Filename: internal/cache/cache.go
package cache

import (
	"container/list"
	"sync"
)

type entry struct {
	key   string
	value []byte
}

// Cache is a thread-safe, in-memory LRU cache.
type Cache struct {
	mu       sync.RWMutex
	capacity int
	ll       *list.List
	items    map[string]*list.Element
}

// New creates a new Cache with a given capacity.
func New(capacity int) *Cache {
	if capacity <= 0 {
		capacity = 1
	}
	return &Cache{
		capacity: capacity,
		ll:       list.New(),
		items:    make(map[string]*list.Element),
	}

}

// Get retrieves a value from the cache.
// It returns the value and a boolean indicating if the key was found.
// If the key is found, it should be marked as recently used.
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.items[key]; ok {

		c.ll.MoveToFront(element)
		return element.Value.(*entry).value, true

	}
	return nil, false
}

// Set adds or updates a key-value pair in the cache.
// If adding the new item exceeds the cache's capacity, it evicts the
// least recently used item.
func (c *Cache) Set(key string, value []byte) {
	// TODO: Implement the Set logic.
	// 1. Acquire a full write lock (c.mu.Lock()).
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.items[key]; ok {
		c.ll.MoveToFront(element)
		element.Value.(*entry).value = value
	} else {
		newEntry := &entry{
			key:   key,
			value: value,
		}

		element := c.ll.PushFront(newEntry)
		c.items[key] = element
	}

	if c.capacity > 0 && c.ll.Len() > c.capacity {
		last := c.ll.Back()
		if last != nil {
			lastEntry := last.Value.(*entry)
			delete(c.items, lastEntry.key)
			c.ll.Remove(last)
		}
	}
}
