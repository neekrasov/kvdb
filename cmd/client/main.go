package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"time"

	"github.com/neekrasov/kvdb/pkg/client"
)

func main() {
	address := flag.String("address", "localhost:3223", "Address of the spider")
	idleTimeout := flag.Duration("idle_timeout", time.Minute, "Idle timeout for connection")
	maxMessageSizeStr := flag.String("max_message_size", "4KB", "Max message size for connection")
	flag.Parse()

	kvdb, err := client.NewClient(&client.KVDBClientConfig{
		Address:        *address,
		IdleTimeout:    *idleTimeout,
		MaxMessageSize: *maxMessageSizeStr,
	})
	if err != nil {
		log.Fatal(err)
	}

	err = kvdb.CLI()
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
}
