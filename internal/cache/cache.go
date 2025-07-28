// Package cache provides a thread-safe, in-memory LRU (Least Recently Used) cache
// with TTL (Time-To-Live) support and automatic background cleanup.
//
// This implementation uses a combination of:
// - HashMap for O(1) key lookups
// - Doubly-linked list for O(1) LRU ordering operations
// - RWMutex for thread-safe concurrent access
// - Background goroutine for automatic expired entry cleanup
//
// Architecture:
//   - The cache maintains a fixed capacity and evicts least recently used items
//   - Each entry can have an optional TTL for automatic expiration
//   - A background "janitor" goroutine runs periodically to clean up expired entries
//   - All operations are thread-safe using read-write mutexes for optimal performance
//
// Usage:
//   cache := cache.New(1000)  // Create cache with capacity of 1000 items
//   defer cache.Stop()        // Always stop the background janitor
//   
//   cache.Set("key", []byte("value"), 5*time.Minute)  // Set with 5min TTL
//   value, found := cache.Get("key")                  // Retrieve value
package cache

import (
	"container/list"
	"sync"
	"time"
)

// entry represents a single cache entry stored in the doubly-linked list.
// Each entry contains the key-value pair and optional expiration time.
//
// Design Notes:
// - key is stored for cleanup operations (could be optimized out in future)
// - value is stored as []byte for flexibility and memory efficiency
// - expiresAt uses time.Time zero value to indicate no expiration
type entry struct {
	key       string    // The cache key (used for cleanup operations)
	value     []byte    // The cached value as byte slice
	expiresAt time.Time // Expiration time (zero value means no expiration)
}

// isExpired checks if this cache entry has expired based on current time.
// Returns false for entries with no expiration (zero time).
//
// Performance: O(1) - simple time comparison
func (e *entry) isExpired() bool {
	// Zero time indicates no expiration
	if e.expiresAt.IsZero() {
		return false
	}

	return time.Now().After(e.expiresAt)
}

// Cache is a thread-safe, in-memory LRU cache with TTL support.
//
// Internal Structure:
// - items: HashMap for O(1) key lookups, maps key -> list.Element
// - ll: Doubly-linked list for O(1) LRU operations (most recent at front)
// - mu: RWMutex for thread-safety (allows concurrent reads)
// - capacity: Maximum number of items (0 = unlimited)
// - Background janitor: Goroutine that periodically cleans expired entries
//
// Thread Safety:
// - All public methods are thread-safe
// - Uses RWMutex to allow concurrent reads while protecting writes
// - Background janitor coordinates with main operations via mutex
//
// Memory Management:
// - LRU eviction when capacity is exceeded
// - Automatic cleanup of expired entries via background janitor
// - Graceful shutdown support to stop background goroutine
type Cache struct {
	// Synchronization
	mu sync.RWMutex // Read-write mutex for thread safety

	// Core data structures
	capacity int                           // Maximum number of items (0 = unlimited)
	ll       *list.List                    // Doubly-linked list for LRU ordering
	items    map[string]*list.Element      // HashMap for O(1) key lookups

	// Background cleanup (janitor goroutine)
	stopJanitor chan struct{} // Signal channel to stop janitor
	janitorDone chan struct{} // Confirmation channel when janitor stops
}

// New creates a new Cache instance with the specified capacity.
//
// Parameters:
//   - capacity: Maximum number of items the cache can hold
//     - If <= 0, defaults to 1 (minimum viable cache)
//     - When capacity is reached, least recently used items are evicted
//
// Behavior:
//   - Initializes all internal data structures
//   - Starts a background janitor goroutine for TTL cleanup
//   - The janitor runs every 1 second by default
//
// Important: Always call cache.Stop() when done to clean up the background goroutine
//
// Example:
//   cache := cache.New(1000)
//   defer cache.Stop()  // Essential for proper cleanup
func New(capacity int) *Cache {
	// Ensure minimum viable capacity
	if capacity <= 0 {
		capacity = 1
	}

	c := &Cache{
		capacity:    capacity,
		ll:          list.New(),                           // Initialize empty doubly-linked list
		items:       make(map[string]*list.Element),       // Initialize empty hashmap
		stopJanitor: make(chan struct{}),                  // Unbuffered channel for stop signal
		janitorDone: make(chan struct{}),                  // Unbuffered channel for completion signal
	}

	// Start background janitor for automatic TTL cleanup
	// Runs every 1 second - could be made configurable in future
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
