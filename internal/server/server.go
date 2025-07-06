package server

import (
	"context"
	"time"

	api "github.com/Devashish08/go-cache-cluster/api/v1"
	"github.com/Devashish08/go-cache-cluster/internal/cache"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// server struct will implement the gRPC server interface.
// It holds a reference to our cache instance.
type server struct {
	// You MUST embed the UnimplementedCacheServer type. This is a technical
	// requirement from gRPC for forward compatibility. It ensures that if we
	// add a new RPC method to our .proto file, older servers don't break.
	api.UnimplementedCacheServer
	cache *cache.Cache
}

// NewGRPCServer creates a new gRPC server instance.
func NewGRPCServer(c *cache.Cache) *server {
	return &server{cache: c}
}

// Set implements the Set RPC method.
func (s *server) Set(ctx context.Context, req *api.SetRequest) (*api.SetResponse, error) {
	duration := time.Duration(req.TtlSeconds) * time.Second
	s.cache.Set(req.Key, req.Value, duration)
	return &api.SetResponse{}, nil
}

// Get implements the Get RPC method.
func (s *server) Get(ctx context.Context, req *api.GetRequest) (*api.GetResponse, error) {
	value, ok := s.cache.Get(req.Key)
	if !ok {
		return nil, status.Error(codes.NotFound, "key not found")
	}

	return &api.GetResponse{Value: value}, nil
}

// Delete implements the Delete RPC method.
func (s *server) Delete(ctx context.Context, req *api.DeleteRequest) (*api.DeleteResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Delete is not yet implemented")
}
