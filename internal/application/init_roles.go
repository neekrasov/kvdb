package application

import (
	"errors"
	"slices"

	"github.com/neekrasov/kvdb/internal/config"
	"github.com/neekrasov/kvdb/internal/database/identity"
	"github.com/neekrasov/kvdb/internal/database/identity/models"
	"github.com/neekrasov/kvdb/internal/database/storage"
	"github.com/neekrasov/kvdb/pkg/logger"
	"go.uber.org/zap"
)

func initRolesStorage(
	storage *storage.Storage,
	cfg *config.Config,
) (*identity.RolesStorage, error) {
	rolesStorage := identity.NewRolesStorage(storage)
	if cfg.Replication != nil && cfg.Replication.ReplicaType == slaveType {
		return rolesStorage, nil
	}

	err := rolesStorage.Save(&models.Role{
		Name: models.RootRoleName,
		Get:  true, Set: true, Del: true,
		Namespace: models.DefaultNameSpace,
	})
	if err != nil {
		logger.Warn("save root role failed", zap.Error(err))
	}

	err = rolesStorage.Save(&models.Role{
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

		err := rolesStorage.Save(&models.Role{
			Name:      role.Name,
			Get:       role.Get,
			Set:       role.Set,
			Del:       role.Del,
			Namespace: role.Namespace,
		})
		if err != nil {
			logger.Warn("save default role failed",
				zap.Error(err), zap.String("name", role.Name),
			)
			continue
		}

		list, err := rolesStorage.Append(role.Name)
		if err != nil {
			logger.Warn("save default role in global list failed", zap.Error(err))
		}

		logger.Debug("created default role", zap.Any("role", role.Name), zap.Strings("list", list))
	}

	return rolesStorage, nil
}
