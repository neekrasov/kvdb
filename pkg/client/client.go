package client

import (
	"errors"
	"fmt"
	"strings"
	"syscall"
	"time"

	"github.com/chzyer/readline"
	"github.com/neekrasov/kvdb/internal/database/command"
	"github.com/neekrasov/kvdb/pkg/client/tcp"
	conversiontypes "github.com/neekrasov/kvdb/pkg/conversion_types"
	sizeparser "github.com/neekrasov/kvdb/pkg/size_parser"
)

var ErrWriteLineFailed = errors.New("write line failed")

// Client - interface for network client.
type Client interface {
	Close() error
	Send(request []byte) ([]byte, error)
}

// KVDBClientConfig holds the configuration settings for the KVDB client.
type KVDBClientConfig struct {
	Username       string        `json:"username"`
	Password       string        `json:"password"`
	Address        string        `json:"address"`        // Address of the KVDB server.
	IdleTimeout    time.Duration `json:"idleTimeout"`    // Idle timeout for the client connection.
	MaxMessageSize string        `json:"maxMessageSize"` // Maximum message size for client communication.
}

// KVDBClient represents a client for interacting with a KVDB server.
type KVDBClient struct {
	client Client // The underlying TCP client used for communication.
}

// NewClient creates and returns a new KVDBClient with the provided configuration.
func NewClient(cfg *KVDBClientConfig) (*KVDBClient, error) {
	if cfg.Address == "" {
		return nil, errors.New("empty address")
	}

	if cfg.Username == "" || cfg.Password == "" {
		return nil, errors.New("username and password must be set")
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

	cmd := fmt.Sprintf("%s %s %s", command.CommandAUTH, cfg.Username, cfg.Password)
	resBytes, err := client.Send([]byte(cmd))
	if err != nil {
		return nil, err
	}

	resString := string(resBytes)
	if strings.Contains(resString, "error:") {
		return nil, errors.New(resString)
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
func (k *KVDBClient) CLI(rl *readline.Instance) error {
	defer func() {
		rl.Close()

		if err := k.client.Close(); err != nil {
			if _, err = rl.Write([]byte(fmt.Sprintf("failed to close client connection: %s", err.Error()))); err != nil {
				return
			}
		}
	}()

	for {
		query, err := rl.ReadSlice()
		if err != nil {
			if errors.Is(err, readline.ErrInterrupt) {
				return nil
			}

			if _, err = rl.Write([]byte(fmt.Sprintf("failed to read stdin: %s", err.Error()))); err != nil {
				return errors.Join(ErrWriteLineFailed, err)
			}
			continue
		}

		if conversiontypes.UnsafeBytesToString(query) == "exit" {
			return nil
		}

		resBytes, err := k.client.Send([]byte(query))
		if err != nil {
			if errors.Is(err, syscall.EPIPE) ||
				errors.Is(err, tcp.ErrTimeout) ||
				errors.Is(err, syscall.ECONNRESET) {
				return err
			}

			if _, err = rl.Write([]byte(fmt.Sprintf("failed to send query: %s", err.Error()))); err != nil {
				return errors.Join(ErrWriteLineFailed, err)
			}
			continue
		}

		if _, err = rl.Write(append(resBytes, '\n')); err != nil {
			return errors.Join(ErrWriteLineFailed, err)
		}
	}
}

// Close - closes all kvdb client connections.
func (k *KVDBClient) Close() error {
	if k.client != nil {
		return k.client.Close()
	}

	return nil
}
