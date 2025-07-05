// Filename: cmd/server/main.go
package main

import (
	"fmt"

	"github.com/Devashish08/go-cache-cluster/internal/cache" // <-- IMPORTANT: Update this path
)

func main() {
	// Create a new cache with a capacity of 2
	c := cache.New(2)
	fmt.Println("Cache created with capacity 2.")

	// Set two items
	fmt.Println("Setting key 'A' with value 'Apple'")
	c.Set("A", []byte("Apple"))
	fmt.Println("Setting key 'B' with value 'Ball'")
	c.Set("B", []byte("Ball"))

	// Get key 'A' to make it the most recently used
	if val, ok := c.Get("A"); ok {
		fmt.Printf("Got key 'A', value: %s. This should make 'A' the most recently used.\n", string(val))
	}

	// Set a third item. This should cause 'B' to be evicted.
	fmt.Println("Setting key 'C' with value 'Cat'. This should evict key 'B'.")
	c.Set("C", []byte("Cat"))

	// --- Verification ---
	fmt.Println("\n--- Verifying cache state ---")

	// Try to get 'B'. It should not be found.
	if _, ok := c.Get("B"); !ok {
		fmt.Println("✅ Key 'B' not found (correctly evicted).")
	} else {
		fmt.Println("❌ FAILED: Key 'B' was found, but it should have been evicted.")
	}

	// Try to get 'A'. It should be found.
	if val, ok := c.Get("A"); ok {
		fmt.Printf("✅ Key 'A' found with value: %s.\n", string(val))
	} else {
		fmt.Println("❌ FAILED: Key 'A' was not found.")
	}

	// Try to get 'C'. It should be found.
	if val, ok := c.Get("C"); ok {
		fmt.Printf("✅ Key 'C' found with value: %s.\n", string(val))
	} else {
		fmt.Println("❌ FAILED: Key 'C' was not found.")
	}
}
