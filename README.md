# Go Cache Cluster

A distributed, in-memory cache system built in Go with gRPC, designed as a learning project to understand distributed systems concepts from first principles.

## 🎯 Project Goals

This project implements a distributed cache system step-by-step, focusing on learning fundamental distributed systems concepts:

- **Phase 1**: Thread-safe LRU cache with TTL support ✅
- **Phase 2**: gRPC API layer ✅  
- **Phase 3**: Consistent hashing for distribution 🚧
- **Phase 4**: Data replication and fault tolerance 📋
- **Phase 5**: Observability and monitoring 📋
- **Phase 6**: Client SDK and persistence 📋

## 🏗️ Current Architecture

```
┌─────────────────┐    gRPC     ┌─────────────────┐
│                 │ ◄────────── │                 │
│  Cache Client   │             │  Cache Server   │
│                 │ ──────────► │                 │
└─────────────────┘             └─────────────────┘
                                         │
                                         ▼
                                ┌─────────────────┐
                                │                 │
                                │   LRU Cache     │
                                │   + TTL         │
                                │   + Janitor     │
                                └─────────────────┘
```

## 🚀 Quick Start

### Prerequisites
- Go 1.21+ 
- Protocol Buffers compiler (for API changes)

### Running the Server
```bash
# Start the cache server (listens on :8080)
go run cmd/server/main.go
```

### Using the Client
```bash
# Set a key-value pair (60s TTL)
go run cmd/client/main.go set user:123 "John Doe"

# Get a value
go run cmd/client/main.go get user:123

# Key will expire after 60 seconds
```

## 📁 Project Structure

```
├── api/v1/                 # Protocol Buffer definitions
│   ├── cache.proto         # gRPC service definition
│   ├── cache.pb.go         # Generated protobuf code
│   └── cache_grpc.pb.go    # Generated gRPC code
├── cmd/
│   ├── server/main.go      # Server entry point
│   └── client/main.go      # CLI client for testing
├── internal/
│   ├── cache/              # Core cache engine
│   │   ├── cache.go        # LRU cache with TTL
│   │   └── cache_test.go   # Cache tests
│   └── server/             # gRPC server implementation
│       └── server.go       # Service handlers
├── .kiro/specs/            # Development specifications
└── go.mod                  # Go module definition
```

## 🔧 Core Components

### Cache Engine (`internal/cache/`)
- **LRU Eviction**: Least Recently Used items are evicted when capacity is reached
- **TTL Support**: Items can expire automatically after a specified duration
- **Thread Safety**: All operations are protected by RWMutex for concurrent access
- **Background Cleanup**: Janitor goroutine removes expired items every second

### gRPC Server (`internal/server/`)
- **Set Operation**: Store key-value pairs with optional TTL
- **Get Operation**: Retrieve values with LRU updates and lazy expiration
- **Delete Operation**: Currently unimplemented (returns Unimplemented error)

### API Definition (`api/v1/`)
- **Protocol Buffers**: Efficient binary serialization
- **gRPC Service**: Type-safe RPC calls
- **Binary Values**: Support for any data type via bytes

## 🧪 Testing

```bash
# Run cache engine tests
go test ./internal/cache/

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

## 📊 Current Features

- ✅ Thread-safe LRU cache
- ✅ TTL with automatic expiration
- ✅ gRPC API (Set, Get)
- ✅ Background cleanup (janitor)
- ✅ Binary-safe values
- ✅ CLI client for testing

## 🚧 Planned Features (Next Phases)

- 🔄 Consistent hashing for key distribution
- 🔄 Multi-node clustering
- 🔄 Data replication for fault tolerance
- 🔄 Cluster membership management
- 🔄 Prometheus metrics
- 🔄 Structured logging
- 🔄 Delete operation implementation
- 🔄 Smart client SDK
- 🔄 Configuration management

## 🎓 Learning Resources

This project follows the learning plan from `plan.md`, which includes:
- Distributed systems concepts
- Consistent hashing algorithms
- CAP theorem and trade-offs
- Replication strategies
- Observability patterns

## 🤝 Development

This is a learning project focused on understanding distributed systems from first principles. Each phase builds upon the previous one, with comprehensive documentation and testing.

### Next Steps
1. Implement consistent hashing (Phase 3)
2. Add request routing between nodes
3. Implement data replication
4. Add cluster membership management
5. Enhance observability

## 📝 License

This project is for educational purposes.