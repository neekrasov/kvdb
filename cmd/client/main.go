package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/chzyer/readline"
	"github.com/neekrasov/kvdb/pkg/client"
)

var (
	ErrWriteLineFailed = errors.New("write line failed")
)

func main() {
	address := flag.String("address", "localhost:3223", "Address of the spider")
	idleTimeout := flag.Duration("idle_timeout", time.Second*10, "Idle timeout for connection")
	maxMessageSizeStr := flag.String("max_message_size", "4KB", "Max message size for connection")
	username := flag.String("username", "", "Username for connection")
	password := flag.String("password", "", "Username for connection")
	flag.Parse()

	kvdb, err := client.New(&client.KVDBClientConfig{
		Address:              *address,
		IdleTimeout:          *idleTimeout,
		MaxMessageSize:       *maxMessageSizeStr,
		Username:             *username,
		Password:             *password,
		MaxReconnectAttempts: 10,
	}, new(client.TCPClientFactory))
	if err != nil {
		log.Fatal(err)
	}

	rl, err := readline.New("$ ")
	if err != nil {
		log.Fatalf("failed to create readline instance: %s", err.Error())
	}

	err = CLI(rl, kvdb)
	if err != nil {
		log.Fatal(err)
	}
}

// CLI runs a command-line interface for interacting with the KVDB client.
func CLI(rl *readline.Instance, client *client.KVDBClient) error {
	defer func() {
		rl.Close()

		if err := client.Close(); err != nil {
			if _, err = rl.Write([]byte(fmt.Sprintf("failed to close client connection: %s", err))); err != nil {
				return
			}
		}
	}()

	for {
		query, err := rl.Readline()
		if err != nil {
			if errors.Is(err, readline.ErrInterrupt) {
				return nil
			}

			if _, err = rl.Write([]byte(fmt.Sprintf("failed to read stdin: %s", err))); err != nil {
				return errors.Join(ErrWriteLineFailed, err)
			}
			continue
		}

		if query == "exit" {
			return nil
		}

		if len(query) == 0 {
			continue
		}

		res, err := client.Send(context.TODO(), query)
		if err != nil {
			if _, err = rl.Write([]byte(fmt.Sprintf("error: sending query failed: %s\n", err))); err != nil {
				return errors.Join(ErrWriteLineFailed, err)
			}

			continue
		}

		if _, err = rl.Write(append([]byte(res), '\n')); err != nil {
			return errors.Join(ErrWriteLineFailed, err)
		}
	}
}
