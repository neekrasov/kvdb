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

func initUserStorage(
	storage *storage.Storage,
	cfg *config.Config,
) (*identity.UsersStorage, error) {
	usersStorage := identity.NewUsersStorage(storage)
	if cfg.Replication != nil && cfg.Replication.ReplicaType == slaveType {
		return usersStorage, nil
	}

	err := usersStorage.SaveRaw(&models.User{
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

		err = usersStorage.SaveRaw(&user)
		if err != nil {
			logger.Warn("save default user failed",
				zap.Error(err),
				zap.String("name", user.Username),
			)
			continue
		}

		list, err := usersStorage.Append(user.Username)
		if err != nil {
			logger.Warn("save default user in global list failed", zap.Error(err))
		}

		user.Password = ""
		logger.Debug("created default user", zap.Any("user", user), zap.Strings("list", list))
	}

	return usersStorage, nil
}
