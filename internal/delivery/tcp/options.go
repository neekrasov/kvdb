package tcp

import "time"

// Default buffer size for reading client data (4 KB).
const defaultBufferSize = 4 << 10

// ServerOption is a functional option type for configuring a Server instance.
type ServerOption func(*Server)

// WithServerIdleTimeout sets the idle timeout for client connections.
func WithServerIdleTimeout(timeout time.Duration) ServerOption {
	return func(server *Server) {
		server.idleTimeout = timeout
	}
}

// WithServerBufferSize sets the buffer size for reading client data.
func WithServerBufferSize(size uint) ServerOption {
	return func(server *Server) {
		server.bufferSize = size
	}
}

// WithServerMaxConnectionsNumber sets the maximum number of concurrent connections.
func WithServerMaxConnectionsNumber(count uint) ServerOption {
	return func(server *Server) {
		server.maxConnections = count
	}
}
