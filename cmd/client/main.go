package main

import (
	"flag"
	"log"
	"time"

	"github.com/chzyer/readline"
	"github.com/neekrasov/kvdb/pkg/client"
)

func main() {
	address := flag.String("address", "localhost:3223", "Address of the spider")
	idleTimeout := flag.Duration("idle_timeout", time.Minute, "Idle timeout for connection")
	maxMessageSizeStr := flag.String("max_message_size", "4KB", "Max message size for connection")
	username := flag.String("username", "", "Username for connection")
	password := flag.String("password", "", "Username for connection")
	flag.Parse()

	kvdb, err := client.NewClient(&client.KVDBClientConfig{
		Address:        *address,
		IdleTimeout:    *idleTimeout,
		MaxMessageSize: *maxMessageSizeStr,
		Username:       *username,
		Password:       *password,
	})
	if err != nil {
		log.Fatal(err)
	}

	rl, err := readline.New("$ ")
	if err != nil {
		log.Fatalf("failed to create readline instance: %s", err.Error())
	}

	err = kvdb.CLI(rl)
	if err != nil {
		log.Fatal(err)
	}
}
