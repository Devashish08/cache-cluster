// Filename: cmd/server/main.go
package main

import (
	"fmt"
	"log"
	"net"

	// Import our packages
	api "github.com/Devashish08/go-cache-cluster/api/v1"
	"github.com/Devashish08/go-cache-cluster/internal/cache"
	"github.com/Devashish08/go-cache-cluster/internal/server"

	// Import gRPC
	"google.golang.org/grpc"
)

func main() {
	port := ":8080"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	c := cache.New(1000)
	defer c.Stop()

	// 2. Create an instance of our gRPC server implementation.
	srv := server.NewGRPCServer(c)

	// 3. Create a new gRPC server from the Go gRPC library.
	grpcServer := grpc.NewServer()

	// 4. Register our server implementation with the gRPC server.
	// This tells the gRPC server how to handle requests for the Cache service.
	api.RegisterCacheServer(grpcServer, srv)

	fmt.Printf("gRPC server listening on %s\n", port)

	// 5. Start the server.
	// This is a blocking call, so the program will run until it's interrupted.
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
