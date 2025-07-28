// Cache Server - Main entry point for the distributed cache server
//
// This is the main application that starts a gRPC server exposing the cache service.
// It demonstrates the basic setup pattern for a gRPC service in Go.
//
// Current Configuration:
//   - Listens on port 8080
//   - Cache capacity: 1000 items
//   - Single-node operation (no clustering)
//   - No authentication or TLS
//
// Architecture Flow:
//   1. Create TCP listener on specified port
//   2. Initialize cache engine with fixed capacity
//   3. Create gRPC server wrapper around cache
//   4. Register service with gRPC framework
//   5. Start serving requests (blocking)
//
// Usage:
//   go run cmd/server/main.go
//   # Server starts on localhost:8080
//   # Use cmd/client/main.go to test
//
// Future Enhancements:
//   - Command-line configuration
//   - Cluster mode support
//   - TLS and authentication
//   - Graceful shutdown handling
//   - Health checks and metrics
package main

import (
	"fmt"
	"log"
	"net"

	// Project imports
	api "github.com/Devashish08/go-cache-cluster/api/v1"
	"github.com/Devashish08/go-cache-cluster/internal/cache"
	"github.com/Devashish08/go-cache-cluster/internal/server"

	// Third-party imports
	"google.golang.org/grpc"
)

func main() {
	// Configuration - TODO: Make these configurable via CLI/config file
	port := ":8080"
	cacheCapacity := 1000

	// Step 1: Create TCP listener
	// This will accept incoming gRPC connections
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to create TCP listener on %s: %v", port, err)
	}

	// Step 2: Initialize cache engine
	// Creates an LRU cache with TTL support and background cleanup
	c := cache.New(cacheCapacity)
	defer c.Stop() // Critical: Stop background janitor on exit

	// Step 3: Create gRPC server implementation
	// This wraps our cache with the gRPC service interface
	srv := server.NewGRPCServer(c)

	// Step 4: Create gRPC server instance
	// This is the actual gRPC framework server
	grpcServer := grpc.NewServer()

	// Step 5: Register our service implementation
	// This tells gRPC how to route Cache service calls to our server
	api.RegisterCacheServer(grpcServer, srv)

	// Step 6: Start serving
	fmt.Printf("🚀 gRPC Cache Server starting on %s\n", port)
	fmt.Printf("📊 Cache capacity: %d items\n", cacheCapacity)
	fmt.Printf("🔧 Ready to accept connections...\n")

	// This is a blocking call - server runs until interrupted
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
