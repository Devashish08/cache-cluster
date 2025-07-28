// Cache Client - Command-line client for testing the cache server
//
// This is a simple CLI client that demonstrates how to interact with the cache server
// using gRPC. It supports basic get and set operations for testing and development.
//
// Supported Commands:
//   - get <key>: Retrieve a value from the cache
//   - set <key> <value>: Store a value in the cache (60s TTL)
//
// Usage Examples:
//
//	go run cmd/client/main.go set mykey "hello world"
//	go run cmd/client/main.go get mykey
//
// Configuration:
//   - Server address: localhost:8080 (hardcoded)
//   - Request timeout: 5 seconds
//   - Default TTL for set operations: 60 seconds
//   - No TLS (insecure connection)
//
// Architecture:
//  1. Parse command-line arguments
//  2. Establish gRPC connection to server
//  3. Create cache client
//  4. Execute requested operation
//  5. Display results
//
// Future Enhancements:
//   - Configurable server address
//   - Support for delete operations
//   - Configurable TTL values
//   - Batch operations
//   - TLS support
//   - Better error handling and user feedback
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	// Project imports
	api "github.com/Devashish08/go-cache-cluster/api/v1"

	// Third-party imports
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Validate command-line arguments
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "❌ Usage: %s <command> <key> [value]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nCommands:\n")
		fmt.Fprintf(os.Stderr, "  get <key>        - Retrieve value for key\n")
		fmt.Fprintf(os.Stderr, "  set <key> <value> - Store key-value pair (60s TTL)\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s set user:123 \"John Doe\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s get user:123\n", os.Args[0])
		os.Exit(1)
	}

	// Configuration - TODO: Make configurable via flags
	serverAddr := "localhost:8080"
	requestTimeout := 5 * time.Second
	defaultTTL := int64(60) // seconds

	// Step 1: Establish gRPC connection
	// Using insecure credentials for development
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("❌ Failed to connect to server %s: %v", serverAddr, err)
	}
	defer conn.Close()

	// Step 2: Create cache client
	client := api.NewCacheClient(conn)

	// Step 3: Set up request context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	// Step 4: Parse command and execute
	command := os.Args[1]
	key := os.Args[2]

	switch command {
	case "get":
		handleGetCommand(ctx, client, key)

	case "set":
		if len(os.Args) < 4 {
			log.Fatalf("❌ SET command requires a value. Usage: set <key> <value>")
		}
		value := os.Args[3]
		handleSetCommand(ctx, client, key, value, defaultTTL)

	default:
		log.Fatalf("❌ Unknown command: %s. Available commands: get, set", command)
	}
}

// handleGetCommand executes a GET operation and displays the result
func handleGetCommand(ctx context.Context, client api.CacheClient, key string) {
	fmt.Printf("🔍 Getting key: %s\n", key)

	res, err := client.Get(ctx, &api.GetRequest{Key: key})
	if err != nil {
		log.Fatalf("❌ GET failed for key '%s': %v", key, err)
	}

	fmt.Printf("✅ GET successful!\n")
	fmt.Printf("   Key: %s\n", key)
	fmt.Printf("   Value: %s\n", string(res.GetValue()))
}

// handleSetCommand executes a SET operation and displays the result
func handleSetCommand(ctx context.Context, client api.CacheClient, key, value string, ttlSeconds int64) {
	fmt.Printf("💾 Setting key: %s = %s (TTL: %ds)\n", key, value, ttlSeconds)

	_, err := client.Set(ctx, &api.SetRequest{
		Key:        key,
		Value:      []byte(value),
		TtlSeconds: ttlSeconds,
	})
	if err != nil {
		log.Fatalf("❌ SET failed for key '%s': %v", key, err)
	}

	fmt.Printf("✅ SET successful!\n")
	fmt.Printf("   Key: %s\n", key)
	fmt.Printf("   Value: %s\n", value)
	fmt.Printf("   TTL: %d seconds\n", ttlSeconds)
}
