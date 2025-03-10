package client

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/neekrasov/kvdb/internal/database"
	"github.com/neekrasov/kvdb/internal/database/compression"
	"github.com/neekrasov/kvdb/internal/database/compute"
	"github.com/neekrasov/kvdb/internal/delivery/tcp"
	"github.com/neekrasov/kvdb/pkg/sizeutil"
)

var (
	ErrMaxReconnects          = errors.New("maximum reconnection attempts reached")
	ErrAuthenticationFailed   = errors.New("authentication failed")
	ErrAuthenticationRequired = errors.New("authentication required")
	ErrInvalidResponseFormat  = errors.New("invalid response format")
)

type (
	// NetClientFactory - interface for creating a new client.
	NetClientFactory interface {
		Make(address string, opts ...tcp.ClientOption) (NetClient, error)
	}

	// NetClient - interface for network client.
	NetClient interface {
		Close() error
		Send(ctx context.Context, request []byte) ([]byte, error)
	}

	Parser interface {
		Parse(query string) (*compute.Command, error)
	}
)

// Config - holds the configuration settings for the KVDB client.
type Config struct {
	Username             string                      `json:"username"`
	Password             string                      `json:"password"`
	Address              string                      `json:"address"`
	MaxMessageSize       string                      `json:"maxMessageSize"`
	MaxReconnectAttempts int                         `json:"maxReconnectAttempts"`
	IdleTimeout          time.Duration               `json:"idleTimeout"`
	ReconnectBaseDelay   time.Duration               `json:"reconnectBaseDelay"`
	Compression          compression.CompressionType `json:"compression"`
}

// Client - represents a client for interacting with a KVDB server.
type Client struct {
	cfg           *Config
	compressor    compression.Compressor
	clientFactory NetClientFactory
	mu            sync.Mutex
	client        NetClient
}

// New - creates and returns a new Client with the provided configuration.
func New(
	ctx context.Context, cfg *Config,
	clientFactory NetClientFactory,
) (*Client, error) {
	if cfg.Address == "" {
		return nil, errors.New("empty address")
	}

	if cfg.Username == "" || cfg.Password == "" {
		return nil, errors.New("username and password must be set")
	}

	if cfg.MaxReconnectAttempts == 0 {
		cfg.MaxReconnectAttempts = 1
	}

	client := &Client{
		cfg:           cfg,
		clientFactory: clientFactory,
	}

	if cfg.Compression != "" {
		compressor, err := compression.New(cfg.Compression)
		if err != nil {
			return nil, err
		}
		client.compressor = compressor
	}

	if err := client.connect(); err != nil {
		return nil, fmt.Errorf("initial connection failed: %w", err)
	}

	if err := client.auth(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	return client, nil
}

// connect - establishes a new connection to the server.
func (k *Client) connect() error {
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
		size, err := sizeutil.ParseSize(k.cfg.MaxMessageSize)
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
func (k *Client) auth(ctx context.Context) error {
	cmd := compute.CommandAUTH.Make(k.cfg.Username, k.cfg.Password)
	res, err := k.client.Send(ctx, []byte(cmd))
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	if strings.Contains(string(res), ErrAuthenticationRequired.Error()) {
		return ErrAuthenticationRequired
	}

	return nil
}

// sendWithRetries - sends a request to the server with retries on failure.
func (k *Client) sendWithRetries(ctx context.Context, request []byte) (string, error) {
	attempt := 0

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		attempt++
		if attempt > k.cfg.MaxReconnectAttempts {
			return "", ErrMaxReconnects
		}

		// TODO: add heartbeats.
		resBytes, err := k.client.Send(ctx, request)
		if err == nil {
			resString := string(resBytes)
			if strings.Contains(resString,
				database.ErrAuthenticationRequired.Error(),
			) {
				if err := k.auth(ctx); err != nil {
					return "", fmt.Errorf("re-authentication failed: %w", err)
				}

				time.Sleep(time.Second)
				continue
			}

			return resString, nil
		}

		if errors.Is(err, tcp.ErrTimeout) {
			continue
		}

		if err := k.reconnect(ctx, attempt); err != nil {
			return "", fmt.Errorf("reconnect failed: %w", err)
		}
	}
}

func (k *Client) send(ctx context.Context, query string) (string, error) {
	res, err := k.sendWithRetries(ctx, []byte(query))
	if err != nil {
		return "", fmt.Errorf("send query failed: %w", err)
	}

	if database.IsError(res) {
		val, ok := database.CutError(res)
		if !ok {
			return "", ErrInvalidResponseFormat
		}

		return "", errors.New(val)
	}

	val, ok := database.CutOK(res)
	if !ok {
		return "", ErrInvalidResponseFormat
	}

	return strings.TrimLeft(val, " "), nil
}

// reconnect - attempts to reconnect with exponential backoff.
func (k *Client) reconnect(ctx context.Context, attempt int) error {
	delay := k.cfg.ReconnectBaseDelay * time.Duration(attempt)

	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return ctx.Err()
	}

	if err := k.connect(); err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}

	if err := k.auth(ctx); err != nil {
		return fmt.Errorf("re-authentication failed: %w", err)
	}

	return nil
}

// Send - sends a query to the KVDB server and returns the result or an error.
func (k *Client) Raw(ctx context.Context, query string) (string, error) {
	return k.send(ctx, query)
}

func (k *Client) Set(ctx context.Context, key, value string) error {
	var data string
	if k.compressor != nil {
		compressed, err := k.compressor.Compress([]byte(value))
		if err != nil {
			return err
		}

		data = base64.StdEncoding.EncodeToString(compressed)
	} else {
		data = value
	}

	if _, err := k.send(ctx, compute.CommandSET.Make(key, data)); err != nil {
		return fmt.Errorf("failed to set '%s': %w", key, err)
	}

	return nil
}

func (k *Client) Get(ctx context.Context, key string) (string, error) {
	response, err := k.send(ctx, compute.CommandGET.Make(key))
	if err != nil {
		return "", fmt.Errorf("failed to get '%s': %w", key, err)
	}

	var value string
	if k.compressor != nil {
		compressedValue, err := base64.StdEncoding.DecodeString(
			strings.TrimSpace(response))
		if err != nil {
			return "", fmt.Errorf("failed to decode base64: %w", err)
		}

		decompressedValue, err := k.compressor.Decompress(compressedValue)
		if err != nil {
			return "", fmt.Errorf("failed to decompress: %w", err)
		}
		value = string(decompressedValue)
	} else {
		value = response
	}

	return value, nil
}

func (k *Client) Del(ctx context.Context, key string) error {
	if _, err := k.send(ctx, compute.CommandDEL.Make(key)); err != nil {
		return fmt.Errorf("failed to del '%s': %w", key, err)
	}

	return nil
}

// Close - closes all kvdb client connections.
func (k *Client) Close() error {
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
