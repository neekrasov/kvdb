package tcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/neekrasov/kvdb/internal/database/identity/models"
	"github.com/neekrasov/kvdb/pkg/logger"
	pkgsync "github.com/neekrasov/kvdb/pkg/sync"
	"go.uber.org/zap"
)

const defaultConnIDLen = 16

type (
	ConnectionID      = string
	Handler           = func(ctx context.Context, sessionID string, request []byte) []byte
	ConnectionHandler = func(ctx context.Context, sessionID string, conn net.Conn) error
)

// Server - a TCP server implementation that handles database queries with connection management and user authentication.
type Server struct {
	listener       net.Listener
	idleTimeout    time.Duration
	semaphore      *pkgsync.Semaphore
	bufferSize     uint
	maxConnections uint

	activeConnections int32
	onconnect         ConnectionHandler
	ondisconnect      ConnectionHandler
}

// NewServer - creates a new instance of the TCP server.
func NewServer(address string, opts ...ServerOption) (*Server, error) {
	if address == "" {
		return nil, errors.New("empty address")
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to start TCP server: %w", err)
	}
	logger.Info("start server listening", zap.String("addr", address))

	server := &Server{
		listener:   listener,
		bufferSize: defaultBufferSize,
	}

	for _, opt := range opts {
		opt(server)
	}

	if mcons := server.maxConnections; mcons > 0 {
		server.semaphore = pkgsync.NewSemaphore(mcons)
	}

	return server, nil
}

// Start - Starts the TCP server listening on the specified address.
func (s *Server) Start(ctx context.Context, handler Handler) {
	if ctx.Err() != nil {
		return
	}

	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) || ctx.Err() != nil {
					logger.Info("server stopped accepting new connections")
					return
				}

				logger.Warn("failed to accept connection", zap.Error(err))
				continue
			}

			sessionID := models.GenSessionID(defaultConnIDLen)
			logger.Debug(
				"accept connection",
				zap.Stringer("remote_addr", conn.RemoteAddr()),
				zap.String("session", sessionID),
			)

			s.semaphore.Acquire()
			atomic.AddInt32(&s.activeConnections, 1)
			go func() {
				defer func() {
					s.semaphore.Release()
					atomic.AddInt32(&s.activeConnections, -1)
				}()

				s.handleConnection(ctx, sessionID, conn, handler)
			}()
		}
	}()

	<-ctx.Done()
}

// handleConnection - manages a single client connection lifecycle
func (s *Server) handleConnection(
	ctx context.Context,
	connID ConnectionID,
	conn net.Conn,
	handler Handler,
) {
	defer func() {
		if v := recover(); v != nil {
			logger.Error(
				"captured panic", zap.Any("panic", v),
				zap.String("stack", string(debug.Stack())),
				zap.String("conn_id", connID))
		}

		if s.ondisconnect != nil {
			if err := s.ondisconnect(ctx, connID, conn); err != nil {
				logger.Warn("executing diconnection handler failed", zap.Error(err))
			}
		}

		if err := conn.Close(); err != nil {
			logger.Warn(
				"failed to close connection",
				zap.Error(err),
				zap.String("conn_id", connID),
			)
		}
		logger.Debug("client disconnected",
			zap.Stringer("address", conn.RemoteAddr()),
			zap.String("conn_id", connID))

	}()

	if s.onconnect != nil {
		if err := s.onconnect(ctx, connID, conn); err != nil {
			logger.Warn("executing connect handler failed", zap.Error(err))
		}
	}

	buffer := make([]byte, s.bufferSize)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if s.idleTimeout != 0 {
			if err := conn.SetDeadline(time.Now().Add(s.idleTimeout)); err != nil {
				logger.Warn("failed to set read/write deadline", zap.Error(err))
				return
			}
		}

		n, err := Read(conn, buffer, int(s.bufferSize))
		if err != nil {
			return
		}

		resp := handler(ctx, connID, buffer[:n])
		if _, err := conn.Write(resp); err != nil {
			logger.Warn("failed to write data", zap.String("conn_id", connID),
				zap.Stringer("address", conn.RemoteAddr()), zap.Error(err))
			return
		}
	}
}

// ActiveConnections - returns the current number of active connections atomically.
func (s *Server) ActiveConnections() int32 {
	return atomic.LoadInt32(&s.activeConnections)
}

func (s *Server) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}

	return nil
}

// Read - reads data from a connection with timeout handling and buffer overflow protection.
func Read(conn net.Conn, b []byte, size int) (int, error) {
	n, err := conn.Read(b)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			logger.Warn("connection timed out", zap.Stringer("remote_addr", conn.RemoteAddr()))
			return 0, err
		}

		if err == io.EOF {
			return 0, nil
		}

		logger.Error("error reading from connection", zap.Error(err))
		return 0, err
	}

	if n == int(size) {
		logger.Warn("buffer overflow", zap.Int("buffer_size_bytes", int(size)))
		return 0, errors.New("buffer overflow")
	}

	return n, nil
}
