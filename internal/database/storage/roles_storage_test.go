package storage_test

import (
	"testing"

	"github.com/neekrasov/kvdb/internal/database/storage"
	"github.com/neekrasov/kvdb/internal/database/storage/models"
	mocks "github.com/neekrasov/kvdb/internal/mocks/storage"
	"github.com/neekrasov/kvdb/pkg/gob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRolesStorage(t *testing.T) {
	mockEngine := mocks.NewEngine(t)
	rolesStorage := storage.NewRolesStorage(mockEngine)

	t.Run("Test Save - success", func(t *testing.T) {
		role := models.Role{Name: "admin"}
		key := storage.MakeKey(models.SystemRoleNameSpace, role.Name)

		mockEngine.On("Get", key).Return("", false).Once()
		mockEngine.On("Set", key, mock.Anything).Return().Once()

		err := rolesStorage.Save(&role)
		assert.NoError(t, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test Save - already exists", func(t *testing.T) {
		role := models.Role{Name: "admin"}
		key := storage.MakeKey(models.SystemRoleNameSpace, role.Name)

		mockEngine.On("Get", key).Return("{}", true).Once()

		err := rolesStorage.Save(&role)
		assert.Equal(t, models.ErrRoleAlreadyExists, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test Get - success", func(t *testing.T) {
		role := models.Role{Name: "user"}
		roleBytes, _ := gob.Encode(role)
		key := storage.MakeKey(models.SystemRoleNameSpace, role.Name)

		mockEngine.On("Get", key).Return(string(roleBytes), true).Once()

		retrievedRole, err := rolesStorage.Get(role.Name)
		assert.NoError(t, err)
		assert.Equal(t, role.Name, retrievedRole.Name)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test Get - not found", func(t *testing.T) {
		roleName := "nonexistent"
		key := storage.MakeKey(models.SystemRoleNameSpace, roleName)

		mockEngine.On("Get", key).Return("", false).Once()

		_, err := rolesStorage.Get(roleName)
		assert.Equal(t, models.ErrRoleNotFound, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test Delete - success", func(t *testing.T) {
		roleName := "temp"
		key := storage.MakeKey(models.SystemRoleNameSpace, roleName)

		mockEngine.On("Get", key).Return("{}", true).Once()
		mockEngine.On("Del", key).Return(nil).Once()

		err := rolesStorage.Delete(roleName)
		assert.NoError(t, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test Delete - not found", func(t *testing.T) {
		roleName := "nonexistent"
		key := storage.MakeKey(models.SystemRoleNameSpace, roleName)

		mockEngine.On("Get", key).Return("", false).Once()

		err := rolesStorage.Delete(roleName)
		assert.Equal(t, models.ErrRoleNotFound, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test Append - success", func(t *testing.T) {
		role := "guest"
		key := models.SystemRolesKey
		listBytes, _ := gob.Encode([]string{})

		mockEngine.On("Get", key).Return(string(listBytes), true).Once()
		mockEngine.On("Set", key, mock.Anything).Return().Once()

		roles, err := rolesStorage.Append(role)
		assert.NoError(t, err)
		assert.Contains(t, roles, role)
		mockEngine.AssertExpectations(t)
	})
}
