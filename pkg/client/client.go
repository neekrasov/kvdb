package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
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
	ErrKeyNotFound            = errors.New("key not found")
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
)

func buildCommandString(
	cmdType compute.CommandType,
	positionalArgs []string,
	namedArgs map[string]string,
) string {
	parts := []string{cmdType.String()}
	parts = append(parts, positionalArgs...)

	for key, value := range namedArgs {
		parts = append(parts, key, value)
	}
	return strings.Join(parts, " ")
}

// Config - holds the configuration settings for the KVDB client.
type Config struct {
	Username             string        `json:"username"`
	Password             string        `json:"password"`
	Address              string        `json:"address"`
	MaxMessageSize       string        `json:"maxMessageSize"`
	Compression          string        `json:"compression"`
	MaxReconnectAttempts int           `json:"maxReconnectAttempts"`
	IdleTimeout          time.Duration `json:"idleTimeout"`
	ReconnectBaseDelay   time.Duration `json:"reconnectBaseDelay"`
	KeepAliveInterval    time.Duration `json:"keepAliveInterval"`
	Namespace            string        `json:"namespace"`
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
	cmd := buildCommandString(compute.CommandAUTH, []string{k.cfg.Username, k.cfg.Password}, nil)
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

		if errors.Is(err, context.Canceled) {
			return "", ctx.Err()
		}

		if err := k.reconnect(ctx, attempt); err != nil {
			return "", fmt.Errorf("reconnect failed: %w", err)
		}
	}
}

func (k *Client) sendRetry(ctx context.Context, query string) (string, error) {
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

// reconnect - attempts to reconnect with lineal backoff.
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
	return k.sendRetry(ctx, query)
}

// Set - stores a value for a given key.
func (k *Client) Set(ctx context.Context, key, value string, opts ...Option) error {
	options := applyOptions(opts)
	processedValue := value

	if options.compressor != nil {
		compressed, err := options.compressor.Compress([]byte(value))
		if err != nil {
			return fmt.Errorf("failed to compress value for key '%s': %w", key, err)
		}
		processedValue = base64.StdEncoding.EncodeToString(compressed)
	}

	args := make(map[string]string)
	if strings.TrimSpace(k.cfg.Namespace) != "" {
		args[compute.NSArg] = k.cfg.Namespace
	}
	if options.namespace != "" {
		args[compute.NSArg] = options.namespace
	}
	if options.ttl != nil {
		args[compute.TTLArg] = options.ttl.String()
	}

	query := buildCommandString(compute.CommandSET, []string{key, processedValue}, args)
	if _, err := k.sendRetry(ctx, query); err != nil {
		return fmt.Errorf("failed to set key '%s': %w", key, err)
	}

	return nil
}

// Get - retrieves the value associated with a given key.
func (k *Client) Get(ctx context.Context, key string, opts ...Option) (string, error) {
	options := applyOptions(opts)

	args := make(map[string]string)
	if k.cfg.Namespace != "" {
		args[compute.NSArg] = k.cfg.Namespace
	}
	if options.namespace != "" {
		args[compute.NSArg] = options.namespace
	}

	query := buildCommandString(compute.CommandGET, []string{key}, args)
	responsePayload, err := k.sendRetry(ctx, query)
	if err != nil {
		if strings.Contains(err.Error(), compute.ErrKeyNotFound.Error()) {
			return "", ErrKeyNotFound
		}

		return "", fmt.Errorf("failed to get key '%s': %w", key, err)
	}

	if options.compressor != nil {
		compressedValue, err := base64.StdEncoding.DecodeString(responsePayload)
		if err != nil {
			return "", fmt.Errorf("failed to decode base64 for key '%s': %w", key, err)
		}
		decompressedValue, err := options.compressor.Decompress(compressedValue)
		if err != nil {
			return "", fmt.Errorf("failed to decompress value for key '%s': %w", key, err)
		}

		return string(decompressedValue), nil
	}

	return responsePayload, nil
}

// Del - removes a key and its value from the storage.
func (k *Client) Del(ctx context.Context, key string, opts ...Option) error {
	options := applyOptions(opts)

	args := make(map[string]string)
	if k.cfg.Namespace != "" {
		args[compute.NSArg] = k.cfg.Namespace
	}
	if options.namespace != "" {
		args[compute.NSArg] = options.namespace
	}

	query := buildCommandString(compute.CommandDEL, []string{key}, args)
	if _, err := k.sendRetry(ctx, query); err != nil {
		return fmt.Errorf("failed to delete key '%s': %w", key, err)
	}

	return nil
}

// Watch - watches the key and returns the value if it has changed.
func (k *Client) Watch(ctx context.Context, key string, opts ...Option) (string, error) {
	options := applyOptions(opts)

	args := make(map[string]string)
	if k.cfg.Namespace != "" {
		args[compute.NSArg] = k.cfg.Namespace
	}
	if options.namespace != "" {
		args[compute.NSArg] = options.namespace
	}

	query := buildCommandString(compute.CommandWATCH, []string{key}, args)
	responsePayload, err := k.sendRetry(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to watch key '%s': %w", key, err)
	}

	if options.compressor != nil {
		compressedValue, err := base64.StdEncoding.DecodeString(responsePayload)
		if err != nil {
			return "", fmt.Errorf("failed to decode base64 for watched key '%s': %w", key, err)
		}
		decompressedValue, err := options.compressor.Decompress(compressedValue)
		if err != nil {
			return "", fmt.Errorf("failed to decompress value for watched key '%s': %w", key, err)
		}

		return string(decompressedValue), nil
	}

	return responsePayload, nil
}

// Stats - returns the collected database statistics.
func (k *Client) Stats(ctx context.Context, key string) (*database.Stats, error) {
	resp, err := k.sendRetry(ctx, compute.CommandSTAT.Make())
	if err != nil {
		return nil, fmt.Errorf("failed to watch key '%s': %w", key, err)
	}

	var stats database.Stats
	if err := json.Unmarshal([]byte(resp), &stats); err != nil {
		return nil, err
	}

	return &stats, nil
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
