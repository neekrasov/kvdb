package application

import (
	"context"
	"errors"
	"slices"

	"github.com/neekrasov/kvdb/internal/database"
	"github.com/neekrasov/kvdb/internal/database/compute"
	"github.com/neekrasov/kvdb/internal/database/storage"
	"github.com/neekrasov/kvdb/internal/database/storage/engine"
	"github.com/neekrasov/kvdb/internal/database/storage/models"
	"github.com/neekrasov/kvdb/internal/database/storage/wal"
	"github.com/neekrasov/kvdb/internal/database/storage/wal/compressor"
	"github.com/neekrasov/kvdb/internal/database/storage/wal/filesystem"
	"github.com/neekrasov/kvdb/internal/database/storage/wal/segment"
	"github.com/neekrasov/kvdb/internal/delivery/tcp"
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

	engine, err := initEngine(a.cfg.Engine)
	if err != nil {
		return err
	}

	wal, err := initWAL(a.cfg.WAL)
	if err != nil {
		return err
	}
	defer func() {
		if err := wal.Close(); err != nil {
			logger.Debug("failed to close wal", zap.Error(err))
		}
	}()

	if wal != nil {
		wal.Start(ctx)
	}

	dstorage, err := storage.NewStorage(engine, storage.WithWALOpt(wal))
	if err != nil {
		return err
	}
	usersStorage, err := initUserStorage(engine, a.cfg)
	if err != nil {
		return err
	}
	rolesStorage, err := initRolesStorage(engine, a.cfg)
	if err != nil {
		return err
	}
	namespaceStorage, err := initNamespacesStorage(engine, a.cfg.DefaultNamespaces)
	if err != nil {
		return err
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

	database := database.New(compute.NewParser(), dstorage, usersStorage, namespaceStorage,
		rolesStorage, storage.NewSessionStorage(), a.cfg.Root)
	server := tcp.NewServer(database, tcpServerOpts...)
	if err := server.Start(ctx, a.cfg.Network.Address); err != nil {
		return err
	}

	return nil
}

func initEngine(cfg *config.EngineConfig) (*engine.Engine, error) {
	if cfg == nil {
		return nil, errors.New("empty engine config")
	}

	return engine.New(engine.WithPartitionNum(cfg.PartitionNum)), nil
}

func initRolesStorage(engine *engine.Engine, cfg *config.Config) (*storage.RolesStorage, error) {
	storage := storage.NewRolesStorage(engine)
	err := storage.Save(&models.Role{
		Name: models.RootRoleName,
		Get:  true, Set: true, Del: true,
		Namespace: models.DefaultNameSpace,
	})
	if err != nil {
		logger.Warn("save root role failed", zap.Error(err))
	}

	err = storage.Save(&models.Role{
		Name: models.DefaultRoleName,
		Get:  true, Set: true, Del: true,
		Namespace: models.DefaultNameSpace,
	})
	if err != nil {
		logger.Warn("save default role failed", zap.Error(err))
	}

	for _, role := range cfg.DefaultRoles {
		if role.Name == "" || role.Name == models.DefaultRoleName {
			return nil, errors.New("invalid role name in default roles")
		}

		contains := slices.ContainsFunc(cfg.DefaultNamespaces, func(nsCfg config.NamespaceConfig) bool {
			return nsCfg.Name == role.Namespace
		})

		if role.Namespace == "" && !contains {
			return nil, errors.New("invalid role namespace in default roles")
		}

		err := storage.Save(&models.Role{
			Name:      role.Name,
			Get:       role.Get,
			Set:       role.Set,
			Del:       role.Del,
			Namespace: role.Namespace,
		})
		if err != nil {
			logger.Warn("save default role failed", zap.Error(err))
		}

		list, err := storage.Append(role.Name)
		if err != nil {
			logger.Warn("save default role in global list failed", zap.Error(err))
		}

		logger.Debug("created default role", zap.Any("role", role.Name), zap.Strings("list", list))
	}

	return storage, nil
}

func initUserStorage(
	engine *engine.Engine,
	cfg *config.Config,
) (*storage.UsersStorage, error) {
	storage := storage.NewUsersStorage(engine)
	err := storage.SaveRaw(&models.User{
		Username: cfg.Root.Username,
		Password: cfg.Root.Password,
		Roles:    []string{models.RootRoleName},
		Cur:      models.DefaultRole,
	})
	if err != nil {
		logger.Warn("save root user failed", zap.Error(err))
	}

	for _, user := range cfg.DefaultUsers {
		if user.Username == "" {
			return nil, errors.New("invalid username in default list")
		}

		if user.Password == "" {
			return nil, errors.New("invalid username in default list")
		}

		if !slices.Contains(user.Roles, models.DefaultRoleName) {
			user.Roles = append(user.Roles, models.DefaultRoleName)
		}

		var userRole config.RoleConfig
		for _, v := range cfg.DefaultRoles {
			if slices.Contains(user.Roles, v.Name) {
				userRole = v
			}
		}

		user := models.User{
			Username: user.Username,
			Password: user.Password,
			Roles:    user.Roles,
			Cur: models.Role{
				Name:      userRole.Name,
				Get:       userRole.Get,
				Set:       userRole.Set,
				Del:       userRole.Del,
				Namespace: userRole.Namespace,
			},
		}

		err = storage.SaveRaw(&user)
		if err != nil {
			logger.Warn("save default user failed", zap.Error(err))
		}

		list, err := storage.Append(user.Username)
		if err != nil {
			logger.Warn("save default user in global list failed", zap.Error(err))
		}

		user.Password = ""
		logger.Debug("created default user", zap.Any("user", user), zap.Strings("list", list))
	}

	return storage, nil
}

func initNamespacesStorage(engine *engine.Engine, defaultNamespaces []config.NamespaceConfig) (*storage.NamespaceStorage, error) {
	storage := storage.NewNamespaceStorage(engine)
	err := storage.Save(models.DefaultNameSpace)
	if err != nil {
		logger.Warn("save default namespace failed", zap.Error(err))
	}

	for _, namespace := range defaultNamespaces {
		if namespace.Name == "" {
			return nil, errors.New("invalid namaspace name in default list")
		}

		err := storage.Save(namespace.Name)
		if err != nil {
			logger.Warn("save namespace in default list failed", zap.Error(err))
		}

		list, err := storage.Append(namespace.Name)
		if err != nil {
			logger.Warn("save default namespace in global list failed", zap.Error(err))
		}

		logger.Debug("created default namespace", zap.Any("namespace", namespace.Name), zap.Strings("list", list))
	}

	return storage, nil
}

func initWAL(cfg *config.WALConfig) (*wal.WAL, error) {
	if cfg == nil {
		return nil, nil
	}

	segmentStorage, err := segment.NewFileSegmentStorage(
		new(filesystem.LocalFileSystem), cfg.DataDir)
	if err != nil {
		return nil, err
	}

	segmentManagerOpts := make([]wal.FileSegmentManagerOpt, 0)
	if cfg.MaxSegmentSize != "" {
		size, err := sizeparser.ParseSize(cfg.MaxSegmentSize)
		if err != nil {
			return nil, err
		}

		segmentManagerOpts = append(segmentManagerOpts, wal.WithMaxSegmentSize(size))
	}
	if cfg.Compression == "gzip" {
		segmentManagerOpts = append(segmentManagerOpts, wal.WithCompressor(new(compressor.GzipCompressor)))
	}

	segmentManager, err := wal.NewFileSegmentManager(
		segmentStorage, segmentManagerOpts...)
	if err != nil {
		return nil, err
	}

	return wal.NewWAL(segmentManager, cfg.FlushingBatchSize, cfg.FlushingBatchTimeout), nil
}
