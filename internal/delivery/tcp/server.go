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

	models "github.com/neekrasov/kvdb/internal/database/identity/models"
	"github.com/neekrasov/kvdb/pkg/logger"
	"github.com/neekrasov/kvdb/pkg/sync"
	"go.uber.org/zap"
)

type contextKey struct{}

var ConnectionStateKey = contextKey{}

type ConnectionState struct {
	User *models.User
}

type Handler = func(ctx context.Context, request []byte) []byte

// Server - a TCP server implementation that handles database queries with connection management and user authentication.
type Server struct {
	listener       net.Listener
	idleTimeout    time.Duration
	bufferSize     uint
	maxConnections uint
	semaphore      *sync.Semaphore

	activeConnections int32
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
		server.semaphore = sync.NewSemaphore(mcons)
	}

	return server, nil
}

// Start - Starts the TCP server listening on the specified address.
func (s *Server) Start(ctx context.Context, handler Handler) {
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
			logger.Debug("accept connection", zap.Stringer("remote_addr", conn.RemoteAddr()))

			s.semaphore.Acquire()
			atomic.AddInt32(&s.activeConnections, 1)
			go func() {
				defer func() {
					s.semaphore.Release()
					atomic.AddInt32(&s.activeConnections, -1)
				}()

				s.handleConnection(ctx, conn, handler)
			}()
		}
	}()

	<-ctx.Done()
}

// handleConnection - manages a single client connection lifecycle
func (s *Server) handleConnection(ctx context.Context, conn net.Conn, handler Handler) {
	defer func() {
		if v := recover(); v != nil {
			logger.Error("captured panic", zap.Any("panic", v), zap.String("stack", string(debug.Stack())))
		}

		if err := conn.Close(); err != nil {
			logger.Warn("failed to close connection", zap.Error(err))
		}

		logger.Debug("client disconnected", zap.Stringer("address", conn.RemoteAddr()))
	}()

	buffer := make([]byte, s.bufferSize)

	state := &ConnectionState{}
	ctx = context.WithValue(ctx, ConnectionStateKey, state)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		n, err := s.read(conn, buffer)
		if err != nil {
			return
		}

		resp := handler(ctx, buffer[:n])
		if _, err := conn.Write(resp); err != nil {
			logger.Warn("failed to write data", zap.Stringer("address", conn.RemoteAddr()), zap.Error(err))
			return
		}

	}
}

// read - reads data from a connection with timeout handling and buffer overflow protection.
func (s *Server) read(conn net.Conn, b []byte) (int, error) {
	if s.idleTimeout != 0 {
		if err := conn.SetReadDeadline(time.Now().Add(s.idleTimeout)); err != nil {
			logger.Warn("failed to set read deadline", zap.Error(err))
			return 0, err
		}
	}

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

	if n == int(s.bufferSize) {
		logger.Warn("buffer overflow", zap.Int("buffer_size_bytes", int(s.bufferSize)))
		return 0, errors.New("buffer overflow")
	}

	return n, nil
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
