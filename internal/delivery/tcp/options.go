package tcp

import "time"

// Default buffer size for reading client data (4 KB).
const defaultBufferSize = 4 << 10

// ServerOption - is a functional option type for configuring a Server instance.
type ServerOption func(*Server)

// WithServerIdleTimeout - sets the idle timeout for client connections.
func WithServerIdleTimeout(timeout time.Duration) ServerOption {
	return func(server *Server) {
		server.idleTimeout = timeout
	}
}

// WithConnectionHandler - handler activates where client connect.
func WithConnectionHandler(handler ConnectionHandler) ServerOption {
	return func(server *Server) {
		server.onconnect = handler
	}
}

// WithDisconnectionHandler - handler activates where client disconnect.
func WithDisconnectionHandler(handler ConnectionHandler) ServerOption {
	return func(server *Server) {
		server.ondisconnect = handler
	}
}

// WithServerBufferSize - sets the buffer size for reading client data.
func WithServerBufferSize(size uint) ServerOption {
	return func(server *Server) {
		server.bufferSize = size
	}
}

// WithServerMaxConnectionsNumber - sets the maximum number of concurrent connections.
func WithServerMaxConnectionsNumber(count uint) ServerOption {
	return func(server *Server) {
		server.maxConnections = count
	}
}

// ClientOption - function type used to configure a Client.
type ClientOption func(*Client)

// WithClientIdleTimeout - sets the idle timeout for the client.
func WithClientIdleTimeout(timeout time.Duration) ClientOption {
	return func(client *Client) {
		client.idleTimeout = timeout
	}
}

// WithClientBufferSize - sets the buffer size for the client.
func WithClientBufferSize(size uint) ClientOption {
	return func(client *Client) {
		client.bufferSize = int(size)
	}
}

// WithKeepAlive - sets the keep alive period for the client.
func WithKeepAlivePeriod(period time.Duration) ClientOption {
	return func(client *Client) {
		client.keepAlivePeriod = period
	}
}
