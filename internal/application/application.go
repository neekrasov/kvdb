package application

import (
	"context"
	"fmt"

	"github.com/neekrasov/kvdb/internal/config"
	"github.com/neekrasov/kvdb/internal/database"
	"github.com/neekrasov/kvdb/internal/database/compute"
	"github.com/neekrasov/kvdb/internal/database/identity"
	"github.com/neekrasov/kvdb/internal/database/storage"
	"github.com/neekrasov/kvdb/internal/database/storage/replication"
	"github.com/neekrasov/kvdb/internal/delivery/tcp"
	"github.com/neekrasov/kvdb/pkg/logger"
	sizeparser "github.com/neekrasov/kvdb/pkg/size_parser"
	"go.uber.org/zap"
)

// Application - represents the main application that starts the server and handles signals.
type Application struct {
	cfg *config.Config
}

// New - creates and returns a new instance of Application.
func New(cfg *config.Config) *Application {
	return &Application{
		cfg: cfg,
	}
}

// Start - initializes configuration, logger, database, and server, then starts the server and handles termination signals.
func (a *Application) Start(ctx context.Context) error {
	logger.InitLogger(a.cfg.Logging.Level, a.cfg.Logging.Output)

	engine, err := initEngine(a.cfg.Engine)
	if err != nil {
		return fmt.Errorf("initialize engine failed: %w", err)
	}

	wal, err := initWAL(a.cfg.WAL)
	if err != nil {
		return fmt.Errorf("initialize wal failed: %w", err)
	}
	defer func() {
		if err := wal.Close(); err != nil {
			logger.Debug("failed to close wal", zap.Error(err))
		}
	}()

	if wal != nil {
		wal.Start(ctx)
	}

	replica, err := initReplica(wal, a.cfg.WAL, a.cfg.Replication)
	if err != nil {
		return fmt.Errorf("initialize replica failed: %w", err)
	}

	master, ok := replica.(*replication.Master)
	if ok {
		go master.Start(ctx)
	}

	slave, ok := replica.(*replication.Slave)
	if ok {
		go slave.Start(ctx)
	}

	var options []storage.StorageOpt
	if wal != nil {
		options = append(options, storage.WithWALOpt(wal))
	}

	if master != nil {
		options = append(options, storage.WithReplicaOpt(master))
	} else if slave != nil {
		options = append(options, storage.WithReplicaOpt(slave))
		options = append(options, storage.WithReplicaStreamOpt(slave.Stream()))
	}

	dstorage, err := storage.NewStorage(engine, options...)
	if err != nil {
		return fmt.Errorf("initialize storage failed: %w", err)
	}
	namespaceStorage, err := initNamespacesStorage(dstorage, a.cfg)
	if err != nil {
		return fmt.Errorf("initialize default namespaces failed: %w", err)
	}
	usersStorage, err := initUserStorage(dstorage, a.cfg)
	if err != nil {
		return fmt.Errorf("initialize default users failed: %w", err)
	}
	rolesStorage, err := initRolesStorage(dstorage, a.cfg)
	if err != nil {
		return fmt.Errorf("initialize default roles failed: %w", err)
	}

	tcpServerOpts := make([]tcp.ServerOption, 0)
	if timeout := a.cfg.Network.IdleTimeout; timeout != 0 {
		logger.Debug("set tcp idle timeout", zap.Stringer("idle_timeout", timeout))
		tcpServerOpts = append(tcpServerOpts, tcp.WithServerIdleTimeout(timeout))
	}

	if mcons := a.cfg.Network.MaxConnections; mcons != 0 {
		logger.Debug("set tcp max connections", zap.Int("max_connections", int(mcons)))
		tcpServerOpts = append(tcpServerOpts, tcp.WithServerMaxConnectionsNumber(mcons))
	}

	if msize := a.cfg.Network.MaxMessageSize; msize != "" {
		size, err := sizeparser.ParseSize(msize)
		if err != nil {
			logger.Error("pase max message size failed", zap.Error(err))
			return err
		}

		logger.Debug("set max_message_size bytes", zap.Int("max_message_size", int(size)))
		tcpServerOpts = append(tcpServerOpts, tcp.WithServerBufferSize(uint(size)))
	}

	db := database.New(
		compute.NewParser(initCommandTrie()), dstorage,
		usersStorage, namespaceStorage, rolesStorage,
		identity.NewSessionStorage(), a.cfg.Root,
	)

	server, err := tcp.NewServer(a.cfg.Network.Address, tcpServerOpts...)
	if err != nil {
		return fmt.Errorf("init tcp server failed: %w", err)
	}

	server.Start(ctx, func(ctx context.Context, request []byte) []byte {
		state, ok := ctx.Value(tcp.ConnectionStateKey).(*tcp.ConnectionState)
		if !ok {
			return []byte("internal server error: connection state not found")
		}

		if state.User == nil {
			user, err := db.Login(string(request))
			if err != nil {
				response := database.WrapError(fmt.Errorf("authentication failed: %w", err))
				return []byte(response)
			}

			state.User = user
			return []byte(database.WrapOK("authentication successful"))
		}

		return []byte(db.HandleQuery(state.User, string(request)))
	})

	if err = server.Close(); err != nil {
		return fmt.Errorf("failed to close server: %w", err)
	}

	return nil
}
