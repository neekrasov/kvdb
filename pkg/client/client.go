package client

import (
	"errors"
	"fmt"
	"syscall"
	"time"

	"github.com/chzyer/readline"
	"github.com/neekrasov/kvdb/pkg/client/tcp"
	conversiontypes "github.com/neekrasov/kvdb/pkg/conversion_types"
	sizeparser "github.com/neekrasov/kvdb/pkg/size_parser"
)

// KVDBClientConfig holds the configuration settings for the KVDB client.
type KVDBClientConfig struct {
	Address        string        `json:"address"`        // Address of the KVDB server.
	IdleTimeout    time.Duration `json:"idleTimeout"`    // Idle timeout for the client connection.
	MaxMessageSize string        `json:"maxMessageSize"` // Maximum message size for client communication.
}

// KVDBClient represents a client for interacting with a KVDB server.
type KVDBClient struct {
	client *tcp.Client // The underlying TCP client used for communication.
}

// NewClient creates and returns a new KVDBClient with the provided configuration.
func NewClient(cfg *KVDBClientConfig) (*KVDBClient, error) {
	if cfg.Address == "" {
		return nil, errors.New("empty address")
	}

	tcpClientOpts := make([]tcp.ClientOption, 0)
	if cfg.IdleTimeout > 0 {
		tcpClientOpts = append(tcpClientOpts, tcp.WithClientIdleTimeout(cfg.IdleTimeout))
	}

	if cfg.MaxMessageSize != "" {
		size, err := sizeparser.ParseSize(cfg.MaxMessageSize)
		if err != nil {
			return nil, fmt.Errorf("parse max message size '%s' failed: %w", cfg.MaxMessageSize, err)
		}
		tcpClientOpts = append(tcpClientOpts, tcp.WithClientBufferSize(uint(size)))
	}

	client, err := tcp.NewClient(cfg.Address, tcpClientOpts...)
	if err != nil {
		return nil, fmt.Errorf("init tcp client failed: %w", err)
	}

	return &KVDBClient{client: client}, nil
}

// Send sends a query to the KVDB server and returns the result or an error.
func (k *KVDBClient) Send(query string) (string, error) {
	resBytes, err := k.client.Send([]byte(query))
	if err != nil {
		return "", err
	}

	return string(resBytes), nil
}

// CLI runs a command-line interface for interacting with the KVDB client.
func (k *KVDBClient) CLI() error {
	rl, err := readline.New("$ ")
	if err != nil {
		return fmt.Errorf("failed to create readline instance: %w", err)
	}

	defer func() {
		rl.Close()

		if err := k.client.Close(); err != nil {
			fmt.Println("failed to close client connection: ", err.Error())
		}
	}()

	for {
		query, err := rl.Readline()
		if err != nil {
			if errors.Is(err, readline.ErrInterrupt) {
				return nil
			}

			fmt.Println("failed to read stdin: ", err.Error())
			continue
		}

		if query == "exit" {
			return nil
		}

		resBytes, err := k.client.Send([]byte(query))
		if err != nil {
			if errors.Is(err, syscall.EPIPE) ||
				errors.Is(err, tcp.ErrTimeout) ||
				errors.Is(err, syscall.ECONNRESET) {
				return err
			}

			fmt.Println("failed to send query: ", err.Error())
			continue
		}

		fmt.Println(conversiontypes.UnsafeBytesToString(resBytes))
	}
}

func (k *KVDBClient) Close() error {
	return k.client.Close()
}
