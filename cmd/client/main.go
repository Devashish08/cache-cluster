package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	api "github.com/Devashish08/go-cache-cluster/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "usage: %s <command> <key> [value]\n", os.Args[0])
		os.Exit(1)
	}

	serverAddr := "localhost:8080"

	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()

	client := api.NewCacheClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	command := os.Args[1]
	key := os.Args[2]

	switch command {
	case "get":
		res, err := client.Get(ctx, &api.GetRequest{Key: key})
		if err != nil {
			log.Fatalf("could not get key '%s': %v", key, err)
		}

		log.Printf("GET successful. Key: %s, Value: %s", key, string(res.GetValue()))

	case "set":
		if len(os.Args) < 4 {
			log.Fatalf("usage: set <key> <value>")
		}
		value := os.Args[3]

		_, err := client.Set(ctx, &api.SetRequest{
			Key:        key,
			Value:      []byte(value),
			TtlSeconds: 60,
		})
		if err != nil {
			log.Fatalf("could not set key '%s': %v", key, err)
		}
		log.Printf("SET successful. Key: %s, Value: %s", key, value)

	default:

		log.Fatalf("unknown command: %s. available commands: get, set", command)
	}
}
