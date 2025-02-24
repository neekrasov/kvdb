package application

import (
	"errors"

	"github.com/neekrasov/kvdb/internal/config"
	"github.com/neekrasov/kvdb/internal/database/identity"
	"github.com/neekrasov/kvdb/internal/database/identity/models"
	"github.com/neekrasov/kvdb/internal/database/storage"
	"github.com/neekrasov/kvdb/pkg/logger"
	"go.uber.org/zap"
)

func initNamespacesStorage(storage *storage.Storage, cfg *config.Config) (*identity.NamespaceStorage, error) {
	nsStorage := identity.NewNamespaceStorage(storage)
	if cfg.Replication != nil && cfg.Replication.ReplicaType == slaveType {
		return nsStorage, nil
	}

	err := nsStorage.Save(models.DefaultNameSpace)
	if err != nil {
		logger.Warn("save default namespace failed",
			zap.Error(err), zap.String("name", models.DefaultNameSpace))
	}

	for _, namespace := range cfg.DefaultNamespaces {
		if namespace.Name == "" {
			return nil, errors.New("invalid namaspace name in default list")
		}

		err := nsStorage.Save(namespace.Name)
		if err != nil {
			logger.Warn("save namespace in default list failed",
				zap.Error(err), zap.String("name", namespace.Name))
			continue
		}

		if _, err := nsStorage.Append(namespace.Name); err != nil {
			logger.Warn("save default namespace in global list failed",
				zap.Error(err), zap.String("name", namespace.Name))
		}

		logger.Debug("created default namespace", zap.Any("namespace", namespace.Name))
	}

	return nsStorage, nil
}
