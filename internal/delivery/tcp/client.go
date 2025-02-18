package tcp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"
)

var (
	ErrTimeout          = errors.New("connection timed out")
	ErrSmallBufferSize  = errors.New("small buffer size")
	ErrConnectionClosed = errors.New("connection closed")
)

func isTimeout(err error) bool {
	netErr, ok := err.(net.Error)
	return ok && netErr.Timeout()
}

// Client represents a TCP client connection.
type Client struct {
	address     string        // Server address.
	connection  net.Conn      // The TCP connection for the client.
	idleTimeout time.Duration // Timeout for idle connection.
	bufferSize  int           // The buffer size for reading data.
}

// NewClient creates a new client with the given address and options.
func NewClient(address string, options ...ClientOption) (*Client, error) {
	client := &Client{
		address:    address,
		bufferSize: defaultBufferSize,
	}

	for _, opt := range options {
		opt(client)
	}

	if err := client.connect(); err != nil {
		return nil, fmt.Errorf("init connection failed: %w", err)
	}

	return client, nil
}

// connect establishes a new connection to the server.
func (c *Client) connect() error {
	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	c.connection = conn

	return nil
}

// Send sends a request to the server and returns the response.
func (c *Client) Send(ctx context.Context, request []byte) ([]byte, error) {
	if c.connection == nil {
		return nil, ErrConnectionClosed
	}

	if len(request) >= c.bufferSize {
		return nil, ErrSmallBufferSize
	}

	if c.idleTimeout > 0 {
		deadline := time.Now().Add(c.idleTimeout)
		if err := c.connection.SetWriteDeadline(deadline); err != nil {
			return nil, fmt.Errorf("failed to set write deadline: %w", err)
		}
		if err := c.connection.SetReadDeadline(deadline); err != nil {
			return nil, fmt.Errorf("failed to set read deadline: %w", err)
		}
	}

	if _, err := c.connection.Write(request); err != nil {
		if isTimeout(err) {
			return nil, errors.Join(ErrTimeout, err)
		}

		return nil, fmt.Errorf("error writing to connection: %w", err)
	}

	response := make([]byte, c.bufferSize)
	n, err := c.connection.Read(response)
	if err != nil {
		if isTimeout(err) {
			return nil, errors.Join(ErrTimeout, err)
		}
		return nil, fmt.Errorf("error reading from connection: %w", err)
	}

	return response[:n], nil
}

// Close closes the client connection.
func (c *Client) Close() error {
	if c.connection != nil {
		if err := c.connection.Close(); err != nil {
			return fmt.Errorf("error closing connection: %w", err)
		}
		c.connection = nil
	}

	return nil
}
