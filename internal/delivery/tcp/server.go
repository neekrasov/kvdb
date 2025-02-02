package tcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/neekrasov/kvdb/pkg/logger"
	"github.com/neekrasov/kvdb/pkg/sync"
	"go.uber.org/zap"
)

type QueryHandler interface {
	HandleQuery(string) string
}

// Server - represents a TCP server that handles client connections and processes queries.
type Server struct {
	database       QueryHandler    // QueryHandler instance for handling queries.
	idleTimeout    time.Duration   // Timeout for idle connections.
	bufferSize     uint            // Maximum size of the read buffer.
	maxConnections uint            // Maximum number of concurrent connections.
	semaphore      *sync.Semaphore // Semaphore to limit concurrent connections.
}

// NewServer - creates a new Server instance with configurable options.
func NewServer(database QueryHandler, opts ...ServerOption) *Server {
	server := &Server{
		database:   database,
		bufferSize: defaultBufferSize,
	}

	for _, opt := range opts {
		opt(server)
	}

	if mcons := server.maxConnections; mcons > 0 {
		server.semaphore = sync.NewSemaphore(mcons)
	}

	return server
}

// Start - begins listening for incoming TCP connections.
func (s *Server) Start(ctx context.Context, address string) error {
	if address == "" {
		return errors.New("empty address")
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to start TCP server: %w", err)
	}

	logger.Info("start server listening", zap.String("addr", address))
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				logger.Warn("failed to accept connection", zap.Error(err))
				return
			}

			logger.Debug("accept connection", zap.Stringer("remote_addr", conn.RemoteAddr()))

			s.semaphore.Acquire()
			go func() {
				defer s.semaphore.Release()
				s.handleConnection(conn)
			}()
		}
	}()

	<-ctx.Done()
	listener.Close()

	return nil
}

// handleConnection - processes an individual client connection.
func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		if v := recover(); v != nil {
			logger.Error("captured panic", zap.Any("panic", v))
		}

		if err := conn.Close(); err != nil {
			logger.Warn("failed to close connection", zap.Error(err))
		}
	}()

	for {
		if s.idleTimeout != 0 {
			if err := conn.SetReadDeadline(time.Now().Add(s.idleTimeout)); err != nil {
				logger.Warn("failed to set read deadline", zap.Error(err))
				return
			}
		}

		buffer := make([]byte, s.bufferSize)
		n, err := conn.Read(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				logger.Warn("connection timed out", zap.Stringer("remote_addr", conn.RemoteAddr()))
				return
			}

			if err == io.EOF {
				return
			}

			logger.Error("error reading from connection", zap.Error(err))
			return
		}

		if n == int(s.bufferSize) {
			logger.Warn("buffer overflow", zap.Int("buffer_size_bytes", int(s.bufferSize)))
			return
		}

		if s.idleTimeout != 0 {
			if err := conn.SetWriteDeadline(time.Now().Add(s.idleTimeout)); err != nil {
				logger.Warn("failed to set write deadline", zap.Error(err))
				return
			}
		}

		response := s.database.HandleQuery(string(buffer[:n]))
		if _, err := conn.Write([]byte(response)); err != nil {
			logger.Warn(
				"failed to write data",
				zap.Stringer("address", conn.RemoteAddr()),
				zap.Error(err),
			)
			return
		}
	}
}
