package application

import (
	"context"

	"github.com/neekrasov/kvdb/internal/compute"
	"github.com/neekrasov/kvdb/internal/database"
	"github.com/neekrasov/kvdb/internal/delivery/tcp"
	"github.com/neekrasov/kvdb/internal/storage"
	"github.com/neekrasov/kvdb/pkg/config"
	"github.com/neekrasov/kvdb/pkg/logger"
	sizeparser "github.com/neekrasov/kvdb/pkg/size_parser"
	"go.uber.org/zap"
)

// Application represents the main application that starts the server and handles signals.
type Application struct {
	cfg *config.Config
}

// New creates and returns a new instance of Application.
func New(cfg *config.Config) *Application {
	return &Application{
		cfg: cfg,
	}
}

// Start initializes configuration, logger, database, and server, then starts the server and handles termination signals.
func (a *Application) Start(ctx context.Context) error {
	logger.InitLogger(a.cfg.Logging.Level, a.cfg.Logging.Output)

	parser := compute.NewParser()
	engine := storage.NewInMemoryEngine()

	tcpServerOpts := make([]tcp.ServerOption, 0)
	if timeout := a.cfg.Network.IdleTimeout; timeout != 0 {
		logger.Debug("set tcp idel timeout", zap.Stringer("idle_timeout", timeout))
		tcpServerOpts = append(tcpServerOpts, tcp.WithServerIdleTimeout(timeout))
	}

	if mcons := a.cfg.Network.MaxConnections; mcons != 0 {
		logger.Debug("set tcp max connections", zap.Int("max_connections", int(mcons)))
		tcpServerOpts = append(tcpServerOpts, tcp.WithServerMaxConnectionsNumber(mcons))
	}

	if msize := a.cfg.Network.MaxMessageSize; msize != "" {
		size, err := sizeparser.ParseSize(msize)
		if err != nil {
			return err
		}

		logger.Debug("set max_message_size bytes", zap.Int("max_message_size", int(size)))
		tcpServerOpts = append(tcpServerOpts, tcp.WithServerBufferSize(uint(size)))
	}

	server := tcp.NewServer(database.New(parser, engine), tcpServerOpts...)
	if err := server.Start(ctx, a.cfg.Network.Address); err != nil {
		return err
	}

	return nil
}
