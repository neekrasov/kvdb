package tcp

import (
	"errors"
	"fmt"
	"net"
	"time"
)

var (
	ErrTimeout         = errors.New("connection timed out")
	ErrSmallBufferSize = errors.New("small buffer size")
)

func isTimeout(err error) bool {
	netErr, ok := err.(net.Error)
	return ok && netErr.Timeout()
}

// Client represents a TCP client connection.
type Client struct {
	connection  net.Conn      // The TCP connection for the client.
	idleTimeout time.Duration // Timeout for idle connection.
	bufferSize  int           // The buffer size for reading data.
}

// NewClient creates a new client with the given address and options.
func NewClient(address string, options ...ClientOption) (*Client, error) {
	connection, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	client := &Client{connection: connection}
	for _, opt := range options {
		opt(client)
	}

	if client.idleTimeout != 0 {
		if err := connection.SetDeadline(time.Now().Add(client.idleTimeout)); err != nil {
			return nil, fmt.Errorf("failed to set deadline for connection: %w", err)
		}
	}

	if client.bufferSize == 0 {
		client.bufferSize = defaultBufferSize
	}

	return client, nil
}

// Send sends a request to the server and returns the response.
func (c *Client) Send(request []byte) ([]byte, error) {
	if _, err := c.connection.Write(request); err != nil {
		if isTimeout(err) {
			return nil, errors.Join(ErrTimeout, err)
		}

		return nil, fmt.Errorf("error writing to connection: %w", err)
	}

	time.Sleep(time.Millisecond)
	response := make([]byte, c.bufferSize)
	n, err := c.connection.Read(response)
	if err != nil {
		return nil, err
	}

	if n == c.bufferSize {
		return nil, ErrSmallBufferSize
	}

	dataStart := 0
	for dataStart < n && response[dataStart] == 0 {
		dataStart++
	}

	if dataStart == n {
		return nil, errors.New("empty message")
	}

	return response, nil
}

// Close closes the client connection.
func (c *Client) Close() error {
	if c.connection != nil {
		return c.connection.Close()
	}

	return nil
}
