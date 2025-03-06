package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/chzyer/readline"
	"github.com/neekrasov/kvdb/internal/database"
	"github.com/neekrasov/kvdb/internal/database/compression"
	"github.com/neekrasov/kvdb/pkg/client"
	"github.com/neekrasov/kvdb/pkg/logger"
	"go.uber.org/zap"
)

var (
	ErrWriteLineFailed = errors.New("write line failed")
)

func main() {
	address := flag.String("address", "localhost:3223", "Address of the server")
	idleTimeout := flag.Duration("idle_timeout", time.Second*10, "Idle timeout for connection")
	maxMessageSizeStr := flag.String("max_message_size", "4KB", "Max message size for connection")
	maxReconnectionAttempts := flag.Int("max_reconnection_attempts", 10, "Max reconnection client attempts")
	username := flag.String("username", "", "Username for connection")
	password := flag.String("password", "", "Username for connection")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	kvdb, err := client.New(ctx,
		&client.Config{
			Address:              *address,
			IdleTimeout:          *idleTimeout,
			MaxMessageSize:       *maxMessageSizeStr,
			Username:             *username,
			Password:             *password,
			MaxReconnectAttempts: *maxReconnectionAttempts,
			Compression:          compression.Zstd,
		}, new(client.TCPClientFactory))
	if err != nil {
		log.Fatal(err)
	}

	rl, err := readline.New("$ ")
	if err != nil {
		log.Fatalf("failed to create readline instance: %s", err.Error())
	}

	if err = CLI(ctx, rl, kvdb); err != nil {
		log.Fatal(err)
	}
}

// CLI runs a command-line interface for interacting with the KVDB client.
func CLI(
	ctx context.Context,
	rl *readline.Instance,
	client *client.Client,
) error {
	defer func() {
		if err := rl.Close(); err != nil {
			logger.Warn("failed to close readline", zap.Error(err))
		}

		if err := client.Close(); err != nil {
			if _, err = rl.Write(fmt.Appendf(nil, "failed to close client connection: %s", err)); err != nil {
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

			if _, err = rl.Write(fmt.Appendf(nil, "failed to read stdin: %s", err)); err != nil {
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

		res, err := client.Raw(ctx, query)
		if err != nil {
			if _, err = rl.Write([]byte(database.WrapError(err) + "\n")); err != nil {
				return errors.Join(ErrWriteLineFailed, err)
			}

			continue
		}

		resBytes := database.WrapOK(string(processInput([]byte(res))))
		if _, err = rl.Write(append([]byte(resBytes), '\n')); err != nil {
			return errors.Join(ErrWriteLineFailed, err)
		}
	}
}

func processInput(input []byte) []byte {
	keyValueBytes, err := jsonToKeyValue(input)
	if err == nil {
		return keyValueBytes
	}

	return input
}

func jsonToKeyValue(data []byte) ([]byte, error) {
	var jsonData any
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, err
	}

	var result bytes.Buffer
	processValue(&result, "", jsonData)
	return result.Bytes(), nil
}

func processValue(result *bytes.Buffer, prefix string, value any) {
	switch v := value.(type) {
	case map[string]any:
		for key, val := range v {
			newPrefix := key
			if prefix != "" {
				newPrefix = fmt.Sprintf("%s.%s", prefix, key)
			}
			processValue(result, newPrefix, val)
		}
	case []any:
		for _, item := range v {
			processValue(result, prefix, item)
		}
	default:
		if prefix == "" {
			result.WriteString(fmt.Sprintf("%v", v))
		} else {
			result.WriteString(fmt.Sprintf("\n%s: %v", prefix, v))
		}
	}
}
