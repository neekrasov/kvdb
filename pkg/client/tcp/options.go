package tcp

import "time"

const defaultBufferSize = 4 << 10 // Default buffer size (4 KB).

// ClientOption is a function type used to configure a Client.
type ClientOption func(*Client)

// WithClientIdleTimeout sets the idle timeout for the client.
func WithClientIdleTimeout(timeout time.Duration) ClientOption {
	return func(client *Client) {
		client.idleTimeout = timeout
	}
}

// WithClientBufferSize sets the buffer size for the client.
func WithClientBufferSize(size uint) ClientOption {
	return func(client *Client) {
		client.bufferSize = int(size)
	}
}
