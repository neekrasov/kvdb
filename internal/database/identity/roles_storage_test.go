package identity_test

import (
	"context"
	"testing"

	"github.com/neekrasov/kvdb/internal/database/identity"
	"github.com/neekrasov/kvdb/internal/database/identity/models"
	"github.com/neekrasov/kvdb/internal/database/storage"
	mocks "github.com/neekrasov/kvdb/internal/mocks/database"
	"github.com/neekrasov/kvdb/pkg/gob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRolesStorage(t *testing.T) {
	t.Parallel()

	mockStorage := mocks.NewStorage(t)
	rolesStorage := identity.NewRolesStorage(mockStorage)

	ctx := context.Background()
	t.Run("Test Save - success", func(t *testing.T) {
		role := models.Role{Name: "admin"}
		key := storage.MakeKey(models.SystemRoleNameSpace, role.Name)

		mockStorage.On("Get", mock.Anything, key).Return("", storage.ErrKeyNotFound).Once()
		mockStorage.On("Set", mock.Anything, key, mock.Anything).Return(nil).Once()

		err := rolesStorage.Save(ctx, &role)
		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test Save - already exists", func(t *testing.T) {
		role := models.Role{Name: "admin"}
		key := storage.MakeKey(models.SystemRoleNameSpace, role.Name)

		mockStorage.On("Get", mock.Anything, key).Return("{}", nil).Once()

		err := rolesStorage.Save(ctx, &role)
		assert.Equal(t, identity.ErrRoleAlreadyExists, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test Get - success", func(t *testing.T) {
		role := models.Role{Name: "user"}
		roleBytes, _ := gob.Encode(role)
		key := storage.MakeKey(models.SystemRoleNameSpace, role.Name)

		mockStorage.On("Get", mock.Anything, key).Return(string(roleBytes), nil).Once()

		retrievedRole, err := rolesStorage.Get(ctx, role.Name)
		assert.NoError(t, err)
		assert.Equal(t, role.Name, retrievedRole.Name)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test Get - not found", func(t *testing.T) {
		roleName := "nonexistent"
		key := storage.MakeKey(models.SystemRoleNameSpace, roleName)

		mockStorage.On("Get", mock.Anything, key).Return("", storage.ErrKeyNotFound).Once()

		_, err := rolesStorage.Get(ctx, roleName)
		assert.Equal(t, identity.ErrRoleNotFound, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test Delete - success", func(t *testing.T) {
		roleName := "temp"
		key := storage.MakeKey(models.SystemRoleNameSpace, roleName)

		mockStorage.On("Get", mock.Anything, key).Return("{}", nil).Once()
		mockStorage.On("Del", mock.Anything, key).Return(nil).Once()

		err := rolesStorage.Delete(ctx, roleName)
		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test Delete - not found", func(t *testing.T) {
		roleName := "nonexistent"
		key := storage.MakeKey(models.SystemRoleNameSpace, roleName)

		mockStorage.On("Get", mock.Anything, key).Return("", storage.ErrKeyNotFound).Once()

		err := rolesStorage.Delete(ctx, roleName)
		assert.Equal(t, identity.ErrRoleNotFound, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test Append - success", func(t *testing.T) {
		role := "guest"
		key := models.SystemRolesKey
		listBytes, _ := gob.Encode([]string{})

		mockStorage.On("Get", mock.Anything, key).Return(string(listBytes), nil).Once()
		mockStorage.On("Set", mock.Anything, key, mock.Anything).Return(nil).Once()

		roles, err := rolesStorage.Append(ctx, role)
		assert.NoError(t, err)
		assert.Contains(t, roles, role)
		mockStorage.AssertExpectations(t)
	})
}
