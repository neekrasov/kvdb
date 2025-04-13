package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/chzyer/readline"
	"github.com/neekrasov/kvdb/internal/database"
	"github.com/neekrasov/kvdb/internal/database/identity"
	"github.com/neekrasov/kvdb/pkg/client"
	"github.com/neekrasov/kvdb/pkg/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	ErrWriteLineFailed = errors.New("write line failed")
	version            = "dev"
	buildTime          = "unknown"
	gitHash            = "unset"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "kvdb-client",
		Short: "KVDB command line client",
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("kvdb client version %s\nbuild time: %s\nhash: %s\n",
				version, buildTime, gitHash)
		},
	})

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the interactive client",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := client.Config{
				Address:              cmd.Flag("address").Value.String(),
				IdleTimeout:          mustParseDuration(cmd.Flag("idle_timeout").Value.String()),
				MaxMessageSize:       cmd.Flag("max_message_size").Value.String(),
				Username:             cmd.Flag("username").Value.String(),
				Password:             cmd.Flag("password").Value.String(),
				MaxReconnectAttempts: mustParseInt(cmd.Flag("max_reconnection_attempts").Value.String()),
				Compression:          cmd.Flag("compression").Value.String(),
				KeepAliveInterval:    mustParseDuration(cmd.Flag("keep_alive").Value.String()),
			}

			ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()

			kvdb, err := client.New(ctx, &cfg, new(client.TCPClientFactory))
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
		},
	}

	runCmd.Flags().StringP("address", "a", "127.0.0.1:3223", "Address of the server")
	runCmd.Flags().Duration("idle_timeout", 0, "Idle timeout for connection")
	runCmd.Flags().Duration("keep_alive", time.Second*2, "Keep alive interval")
	runCmd.Flags().String("compression", "", "Type for message compression (gzip, zstd, flate, bzip2)")
	runCmd.Flags().String("max_message_size", "4KB", "Max message size for connection")
	runCmd.Flags().Int("max_reconnection_attempts", 10, "Max reconnection client attempts")
	runCmd.Flags().String("username", "", "Username for connection")
	runCmd.Flags().String("password", "", "Password for connection")

	rootCmd.AddCommand(runCmd)
	if err := rootCmd.Execute(); err != nil {
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

		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()

		res, err := client.Raw(ctx, query)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				continue
			}

			if strings.Contains(err.Error(), identity.ErrExpiresSession.Error()) {
				if _, err = rl.Write([]byte(database.WrapError(identity.ErrExpiresSession) + "\n")); err != nil {
					return errors.Join(ErrWriteLineFailed, err)
				}
				return nil
			}

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
			result.WriteString(fmt.Sprintf("%v ", v))
		} else {
			result.WriteString(fmt.Sprintf("\n%s: %v", prefix, v))
		}
	}
}

func mustParseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	return d
}

func mustParseInt(s string) int {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	if err != nil {
		return 0
	}
	return i
}
