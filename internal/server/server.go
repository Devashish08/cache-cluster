// Package server provides gRPC server implementation for the cache service.
//
// This package implements the Cache service defined in the protobuf specification,
// providing a network interface to the underlying cache engine. It handles:
// - gRPC request/response conversion
// - Error handling and status codes
// - Integration with the cache engine
//
// Architecture:
//   - server struct implements the generated CacheServer interface
//   - Each RPC method delegates to the underlying cache operations
//   - Proper gRPC error codes are returned for different scenarios
//   - Thread-safety is handled by the underlying cache implementation
//
// Current Status:
//   - Set and Get operations are fully implemented
//   - Delete operation is stubbed (returns Unimplemented)
//   - Single-node operation only (no clustering yet)
//
// Future Enhancements:
//   - Distributed request routing
//   - Replication support
//   - Metrics and observability
//   - Authentication and authorization
package server

import (
	"context"
	"time"

	api "github.com/Devashish08/go-cache-cluster/api/v1"
	"github.com/Devashish08/go-cache-cluster/internal/cache"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// server implements the gRPC CacheServer interface.
//
// Design Notes:
// - Embeds UnimplementedCacheServer for forward compatibility
// - Holds a reference to the cache engine for actual storage operations
// - All methods are thread-safe due to underlying cache thread-safety
//
// Thread Safety:
// - The server itself is stateless and thread-safe
// - All state is managed by the underlying cache instance
// - Multiple goroutines can safely call server methods concurrently
type server struct {
	// UnimplementedCacheServer provides forward compatibility.
	// If new RPC methods are added to the .proto file, older servers
	// won't break - they'll return "Unimplemented" for new methods.
	api.UnimplementedCacheServer
	
	// cache is the underlying storage engine
	// All RPC operations delegate to this cache instance
	cache *cache.Cache
}

// NewGRPCServer creates a new gRPC server instance with the given cache.
//
// Parameters:
//   - c: The cache instance to use for storage operations
//
// Returns:
//   - A server instance ready to handle gRPC requests
//
// Usage:
//   cache := cache.New(1000)
//   server := server.NewGRPCServer(cache)
//   // Register with gRPC and start serving...
func NewGRPCServer(c *cache.Cache) *server {
	return &server{cache: c}
}

// Set implements the Set RPC method from the Cache service.
//
// This method stores a key-value pair in the cache with an optional TTL.
//
// Parameters:
//   - ctx: Request context (currently unused but required by gRPC)
//   - req: SetRequest containing key, value, and TTL in seconds
//
// Behavior:
//   - Converts TTL from seconds to time.Duration
//   - TTL of 0 means no expiration
//   - Overwrites existing keys
//   - May trigger LRU eviction if cache is at capacity
//
// Returns:
//   - Empty SetResponse on success
//   - Never returns an error (cache.Set cannot fail)
//
// Thread Safety: Safe for concurrent use
func (s *server) Set(ctx context.Context, req *api.SetRequest) (*api.SetResponse, error) {
	// Convert TTL from seconds to Go duration
	// Zero duration means no expiration
	duration := time.Duration(req.TtlSeconds) * time.Second
	
	// Delegate to cache engine - this operation cannot fail
	s.cache.Set(req.Key, req.Value, duration)
	
	return &api.SetResponse{}, nil
}

// Get implements the Get RPC method from the Cache service.
//
// This method retrieves a value from the cache by key.
//
// Parameters:
//   - ctx: Request context (currently unused but required by gRPC)
//   - req: GetRequest containing the key to retrieve
//
// Behavior:
//   - Performs lazy expiration check (expired entries are removed)
//   - Updates LRU ordering (accessed key becomes most recent)
//   - Returns NotFound error if key doesn't exist or has expired
//
// Returns:
//   - GetResponse with the value if key exists and is not expired
//   - gRPC NotFound error if key doesn't exist or has expired
//
// Thread Safety: Safe for concurrent use
func (s *server) Get(ctx context.Context, req *api.GetRequest) (*api.GetResponse, error) {
	// Attempt to retrieve from cache
	// This also handles lazy expiration and LRU updates
	value, found := s.cache.Get(req.Key)
	
	if !found {
		// Return proper gRPC error code for missing keys
		// This covers both non-existent keys and expired keys
		return nil, status.Error(codes.NotFound, "key not found")
	}

	return &api.GetResponse{Value: value}, nil
}

// Delete implements the Delete RPC method from the Cache service.
//
// Current Status: NOT IMPLEMENTED
//
// This method should remove a key from the cache but is currently stubbed.
// It returns an Unimplemented error to indicate the feature is not ready.
//
// TODO: Implement actual deletion logic
//   - Add Delete method to cache.Cache
//   - Remove from both hashmap and linked list
//   - Handle thread safety
//   - Return appropriate success/error responses
//
// Parameters:
//   - ctx: Request context (unused)
//   - req: DeleteRequest containing the key to delete
//
// Returns:
//   - Currently always returns Unimplemented error
//   - Future: DeleteResponse on success, NotFound if key doesn't exist
func (s *server) Delete(ctx context.Context, req *api.DeleteRequest) (*api.DeleteResponse, error) {
	// TODO: Implement delete functionality
	// This is part of the planned enhancements
	return nil, status.Error(codes.Unimplemented, "Delete is not yet implemented")
}
