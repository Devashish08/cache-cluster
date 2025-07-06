// Filename: internal/cache/cache.go
package cache

import (
	"container/list"
	"sync"
	"time"
)

type entry struct {
	key       string
	value     []byte
	expiresAt time.Time
}

func (e *entry) isExpired() bool {
	if e.expiresAt.IsZero() {
		return false
	}

	return time.Now().After(e.expiresAt)
}

// Cache is a thread-safe, in-memory LRU cache.
type Cache struct {
	mu       sync.RWMutex
	capacity int
	ll       *list.List
	items    map[string]*list.Element

	stopJanitor chan struct{}
	janitorDone chan struct{}
}

// New creates a new Cache with a given capacity.
func New(capacity int) *Cache {
	if capacity <= 0 {
		capacity = 1
	}
	c := &Cache{
		capacity:    capacity,
		ll:          list.New(),
		items:       make(map[string]*list.Element),
		stopJanitor: make(chan struct{}),
		janitorDone: make(chan struct{}),
	}

	go c.runJanitor(1 * time.Second)
	return c

}

func (c *Cache) runJanitor(interval time.Duration) {
	defer close(c.janitorDone)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.evictExpired()
		case <-c.stopJanitor:
			return
		}
	}
}

func (c *Cache) evictExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	var toRemove []*list.Element

	for element := c.ll.Front(); element != nil; element = element.Next() {
		ent := element.Value.(*entry)
		if ent.isExpired() {
			toRemove = append(toRemove, element)
		}
	}

	for _, element := range toRemove {
		ent := element.Value.(*entry)
		c.ll.Remove(element)
		delete(c.items, ent.key)
	}

}

func (c *Cache) Stop() {
	select {
	case c.stopJanitor <- struct{}{}:
		<-c.janitorDone
	default:
	}
}

// Get retrieves a value from the cache.
// It returns the value and a boolean indicating if the key was found.
// If the key is found, it should be marked as recently used.
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.items[key]; ok {
		ent := element.Value.(*entry)
		if ent.isExpired() {
			c.ll.Remove(element)
			delete(c.items, ent.key)
			return nil, false
		}
		c.ll.MoveToFront(element)
		return ent.value, true

	}
	return nil, false
}

// Set adds or updates a key-value pair in the cache.
// If adding the new item exceeds the cache's capacity, it evicts the
// least recently used item.
func (c *Cache) Set(key string, value []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	if element, ok := c.items[key]; ok {
		c.ll.MoveToFront(element)
		ent := element.Value.(*entry)
		ent.value = value
		ent.expiresAt = expiresAt
	} else {
		newEntry := &entry{
			key:       key,
			value:     value,
			expiresAt: expiresAt,
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

func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ll.Len()
}
