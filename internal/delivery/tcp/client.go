package tcp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
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

// Client - represents a TCP client connection.
type Client struct {
	address         string        // Server address.
	idleTimeout     time.Duration // Timeout for idle connection.
	bufferSize      int           // The buffer size for reading data.
	keepAlivePeriod time.Duration // Period for keep alive

	mu         sync.Mutex
	connection net.Conn // The TCP connection for the client.
}

// NewClient - creates a new client with the given address and options.
func NewClient(address string, options ...ClientOption) (*Client, error) {
	client := &Client{
		address:    address,
		bufferSize: defaultBufferSize,
	}

	for _, opt := range options {
		opt(client)
	}

	if client.keepAlivePeriod == 0 {
		client.keepAlivePeriod = time.Second
	}

	if err := client.Сonnect(); err != nil {
		return nil, fmt.Errorf("init connection failed: %w", err)
	}

	return client, nil
}

// Сonnect - establishes a new connection to the server.
func (c *Client) Сonnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	c.connection = conn

	tcpConn := conn.(*net.TCPConn)
	if err := tcpConn.SetKeepAlive(true); err != nil {
		return fmt.Errorf("setting keep alive failed: %w", err)
	}

	if err := tcpConn.SetKeepAlivePeriod(c.keepAlivePeriod); err != nil {
		return fmt.Errorf("setting keep alive period failed: %w", err)
	}

	return nil
}

// Send - sends a request to the server and returns the response.
func (c *Client) Send(ctx context.Context, request []byte) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connection == nil {
		return nil, ErrConnectionClosed
	}

	if c.idleTimeout > 0 {
		deadline := time.Now().Add(c.idleTimeout)
		if err := c.connection.SetDeadline(deadline); err != nil {
			return nil, fmt.Errorf("failed to set write deadline: %w", err)
		}
	}

	if _, err := c.connection.Write(request); err != nil {
		if isTimeout(err) {
			return nil, errors.Join(ErrTimeout, err)
		}

		return nil, fmt.Errorf("error writing to connection: %w", err)
	}

	var readErr error
	done := make(chan struct{})
	response := make([]byte, c.bufferSize)

	go func() {
		n, err := c.connection.Read(response)
		if err != nil {
			if isTimeout(err) {
				readErr = errors.Join(ErrTimeout, err)
			} else {
				readErr = fmt.Errorf("error reading from connection: %w", err)
			}
		} else if n >= c.bufferSize {
			readErr = ErrSmallBufferSize
		} else {
			response = response[:n]
		}
		close(done)
	}()

	select {
	case <-done:
		return response, readErr
	case <-ctx.Done():
		if err := c.cancelCurrentOperationLocked(); err != nil {
			return nil, fmt.Errorf("failed to cancel operation: %w", err)
		}

		return nil, fmt.Errorf("operation canceled: %w", ctx.Err())
	}
}

func (c *Client) cancelCurrentOperationLocked() error {
	if _, err := c.connection.Write([]byte(cancelCommand)); err != nil {
		return fmt.Errorf("failed to send cancel request: %w", err)
	}

	return nil
}

// Close - closes the client connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connection != nil {
		if err := c.connection.Close(); err != nil {
			return fmt.Errorf("error closing connection: %w", err)
		}
		c.connection = nil
	}

	return nil
}
