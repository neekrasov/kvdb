package client

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/neekrasov/kvdb/internal/database"
	"github.com/neekrasov/kvdb/internal/database/compute"
	"github.com/neekrasov/kvdb/internal/delivery/tcp"
	sizeparser "github.com/neekrasov/kvdb/pkg/size_parser"
)

var (
	ErrMaxReconnects          = errors.New("maximum reconnection attempts reached")
	ErrAuthenticationFailed   = errors.New("authentication failed")
	ErrAuthenticationRequired = errors.New("authentication required")
)

// ClientFactory - interface for creating a new client.
type ClientFactory interface {
	Make(address string, opts ...tcp.ClientOption) (Client, error)
}

// Client - interface for network client.
type Client interface {
	Close() error
	Send(ctx context.Context, request []byte) ([]byte, error)
}

// KVDBClientConfig - holds the configuration settings for the KVDB client.
type KVDBClientConfig struct {
	Username             string        `json:"username"`
	Password             string        `json:"password"`
	Address              string        `json:"address"`
	MaxMessageSize       string        `json:"maxMessageSize"`
	MaxReconnectAttempts int           `json:"maxReconnectAttempts"`
	IdleTimeout          time.Duration `json:"idleTimeout"`
	ReconnectBaseDelay   time.Duration `json:"reconnectBaseDelay"`
}

// KVDBClient - represents a client for interacting with a KVDB server.
type KVDBClient struct {
	clientFactory ClientFactory
	cfg           *KVDBClientConfig
	mu            sync.Mutex
	client        Client
}

// New - creates and returns a new KVDBClient with the provided configuration.
func New(cfg *KVDBClientConfig, clientFactory ClientFactory) (*KVDBClient, error) {
	if cfg.Address == "" {
		return nil, errors.New("empty address")
	}

	if cfg.Username == "" || cfg.Password == "" {
		return nil, errors.New("username and password must be set")
	}

	if cfg.MaxReconnectAttempts == 0 {
		cfg.MaxReconnectAttempts = 1
	}

	kvdbClient := &KVDBClient{
		cfg:           cfg,
		clientFactory: clientFactory,
	}

	if err := kvdbClient.connect(); err != nil {
		return nil, fmt.Errorf("initial connection failed: %w", err)
	}

	if err := kvdbClient.auth(); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	return kvdbClient, nil
}

// connect - establishes a new connection to the server.
func (k *KVDBClient) connect() error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.client != nil {
		_ = k.client.Close()
	}

	tcpClientOpts := make([]tcp.ClientOption, 0)
	if k.cfg.IdleTimeout > 0 {
		tcpClientOpts = append(tcpClientOpts, tcp.WithClientIdleTimeout(k.cfg.IdleTimeout))
	}

	if k.cfg.MaxMessageSize != "" {
		size, err := sizeparser.ParseSize(k.cfg.MaxMessageSize)
		if err != nil {
			return fmt.Errorf("parse max message size '%s' failed: %w", k.cfg.MaxMessageSize, err)
		}
		tcpClientOpts = append(tcpClientOpts, tcp.WithClientBufferSize(uint(size)))
	}

	client, err := k.clientFactory.Make(k.cfg.Address, tcpClientOpts...)
	if err != nil {
		return err
	}
	k.client = client

	return nil
}

// auth - performs authentication with the server.
func (k *KVDBClient) auth() error {
	cmd := fmt.Sprintf("%s %s %s", compute.CommandAUTH, k.cfg.Username, k.cfg.Password)
	res, err := k.client.Send(context.Background(), []byte(cmd))
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	if strings.Contains(string(res), ErrAuthenticationRequired.Error()) {
		return ErrAuthenticationRequired
	}

	return nil
}

// sendWithRetries - sends a request to the server with retries on failure.
func (k *KVDBClient) sendWithRetries(ctx context.Context, request []byte) (string, error) {
	attempt := 0

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			attempt++
			if attempt > k.cfg.MaxReconnectAttempts {
				return "", ErrMaxReconnects
			}

			resBytes, err := k.client.Send(ctx, request)
			if err == nil {
				resString := string(resBytes)
				if strings.Contains(resString, database.ErrAuthenticationRequired.Error()) {
					if err := k.auth(); err != nil {
						return "", fmt.Errorf("re-authentication failed: %w", err)
					}
					continue
				}

				return resString, nil
			}

			if err := k.reconnect(ctx, attempt); err != nil {
				return "", fmt.Errorf("reconnect failed: %w", err)
			}
		}
	}
}

// reconnect - attempts to reconnect with exponential backoff.
func (k *KVDBClient) reconnect(ctx context.Context, attempt int) error {
	delay := k.cfg.ReconnectBaseDelay * time.Duration(attempt)

	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return ctx.Err()
	}

	if err := k.connect(); err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}

	if err := k.auth(); err != nil {
		return fmt.Errorf("re-authentication failed: %w", err)
	}

	return nil
}

// Send - sends a query to the KVDB server and returns the result or an error.
func (k *KVDBClient) Send(ctx context.Context, query string) (string, error) {
	res, err := k.sendWithRetries(ctx, []byte(query))
	if err != nil {
		return "", fmt.Errorf("send failed: %w", err)
	}

	return res, nil
}

// Close - closes all kvdb client connections.
func (k *KVDBClient) Close() error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.client != nil {
		if err := k.client.Close(); err != nil {
			return fmt.Errorf("error closing connection: %w", err)
		}
		k.client = nil
	}

	return nil
}
