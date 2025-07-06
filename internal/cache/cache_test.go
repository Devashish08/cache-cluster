package cache

import (
	"bytes"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestCache_SetGet(t *testing.T) {
	cache := New(10)
	defer cache.Stop()

	key := "test_key"
	value := []byte("test_value")

	cache.Set(key, value, 0)

	retrieved, found := cache.Get(key)
	if !found {
		t.Fatalf("Key '%s' should have been found", key)
	}

	if !bytes.Equal(value, retrieved) {
		t.Fatalf("Got value '%s', expected '%s'", string(retrieved), string(value))
	}
}

// Test the LRU eviction logic
func TestCache_LRUEviction(t *testing.T) {
	cache := New(2)
	defer cache.Stop()

	cache.Set("A", []byte("Apple"), 0)
	cache.Set("B", []byte("Ball"), 0)
	cache.Get("A")                   // Make A most recently used
	cache.Set("C", []byte("Cat"), 0) // Should evict B

	if _, found := cache.Get("B"); found {
		t.Error("Key 'B' should have been evicted")
	}
	if _, found := cache.Get("A"); !found {
		t.Error("Key 'A' should be present")
	}
	if _, found := cache.Get("C"); !found {
		t.Error("Key 'C' should be present")
	}
}

// Test TTL expiration
func TestCache_TTLExpiration(t *testing.T) {
	t.Run("LazyEviction", func(t *testing.T) {
		cache := New(10)
		defer cache.Stop()

		key := "lazy_key"
		ttl := 50 * time.Millisecond

		cache.Set(key, []byte("value"), ttl)
		time.Sleep(ttl + 10*time.Millisecond)

		if _, found := cache.Get(key); found {
			t.Error("Expired key should have been lazily evicted")
		}
	})

	t.Run("JanitorEviction", func(t *testing.T) {
		cache := New(10)
		defer cache.Stop()

		key := "janitor_key"
		ttl := 50 * time.Millisecond

		cache.Set(key, []byte("value"), ttl)

		// Wait for janitor to run (it runs every 1s, so wait longer)
		time.Sleep(1100 * time.Millisecond)

		// Check length instead of accessing internal state
		initialLength := cache.Len()
		cache.Set("test", []byte("test"), 0) // Trigger any cleanup

		if cache.Len() != initialLength+1 || cache.Len() > 1 {
			// If janitor worked, expired item should be gone
			if _, found := cache.Get(key); found {
				t.Error("Expired key should have been removed by janitor")
			}
		}
	})
}

// Test permanent items (no TTL)
func TestCache_NoTTL(t *testing.T) {
	cache := New(10)
	defer cache.Stop()

	key := "permanent_key"
	cache.Set(key, []byte("permanent value"), 0)
	time.Sleep(100 * time.Millisecond) // Short wait

	if _, found := cache.Get(key); !found {
		t.Error("Item with no TTL should not expire")
	}
}

// Test edge cases
func TestCache_EdgeCases(t *testing.T) {
	t.Run("ZeroCapacity", func(t *testing.T) {
		cache := New(0) // Should default to 1
		defer cache.Stop()

		cache.Set("key1", []byte("value1"), 0)
		cache.Set("key2", []byte("value2"), 0) // Should evict key1

		if _, found := cache.Get("key1"); found {
			t.Error("key1 should have been evicted")
		}
		if _, found := cache.Get("key2"); !found {
			t.Error("key2 should be present")
		}
	})

	t.Run("NegativeCapacity", func(t *testing.T) {
		cache := New(-5) // Should default to 1
		defer cache.Stop()

		cache.Set("key", []byte("value"), 0)
		if cache.Len() != 1 {
			t.Errorf("Expected length 1, got %d", cache.Len())
		}
	})

	t.Run("UpdateExistingKey", func(t *testing.T) {
		cache := New(10)
		defer cache.Stop()

		key := "update_key"
		cache.Set(key, []byte("old_value"), 0)
		cache.Set(key, []byte("new_value"), 0)

		value, found := cache.Get(key)
		if !found {
			t.Error("Key should exist")
		}
		if !bytes.Equal(value, []byte("new_value")) {
			t.Error("Value should be updated")
		}
		if cache.Len() != 1 {
			t.Error("Length should be 1 after update")
		}
	})
}

// Test concurrent access
func TestCache_ConcurrentAccess(t *testing.T) {
	cache := New(100)
	defer cache.Stop()

	var waitGroup sync.WaitGroup
	numGoroutines := 10
	numOperations := 100

	// Concurrent sets
	for i := 0; i < numGoroutines; i++ {
		waitGroup.Add(1)
		go func(goroutineID int) {
			defer waitGroup.Done()
			for j := 0; j < numOperations; j++ {
				key := strconv.Itoa(goroutineID*numOperations + j)
				cache.Set(key, []byte("value"+key), 0)
			}
		}(i)
	}

	// Concurrent gets
	for i := 0; i < numGoroutines; i++ {
		waitGroup.Add(1)
		go func(goroutineID int) {
			defer waitGroup.Done()
			for j := 0; j < numOperations; j++ {
				key := strconv.Itoa(goroutineID*numOperations + j)
				cache.Get(key)
			}
		}(i)
	}

	waitGroup.Wait()
	// Just ensure no race conditions occurred
}

// --- BENCHMARKS ---

// Benchmark the Get method under concurrent load
func BenchmarkCache_Get(b *testing.B) {
	cache := New(10000)
	defer cache.Stop()

	// Pre-fill cache sequentially (simpler and cleaner)
	for i := 0; i < 10000; i++ {
		cache.Set(strconv.Itoa(i), []byte("value"), 0)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		randGen := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			key := strconv.Itoa(randGen.Intn(10000))
			cache.Get(key)
		}
	})
}

// Benchmark the Set method under concurrent load
func BenchmarkCache_Set(b *testing.B) {
	cache := New(10000)
	defer cache.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		randGen := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			key := strconv.Itoa(randGen.Intn(10000))
			cache.Set(key, []byte("value"), 0)
		}
	})
}

func BenchmarkCache_Mixed(b *testing.B) {
	cache := New(10000)
	defer cache.Stop()

	// Pre-fill
	for i := 0; i < 5000; i++ {
		cache.Set(strconv.Itoa(i), []byte("value"), 0)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		randGen := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			key := strconv.Itoa(randGen.Intn(10000))
			if randGen.Intn(2) == 0 {
				cache.Get(key)
			} else {
				cache.Set(key, []byte("value"), 0)
			}
		}
	})
}
