package identity_test

import (
	"testing"

	"github.com/neekrasov/kvdb/internal/database/identity"
	"github.com/neekrasov/kvdb/internal/database/identity/models"
	"github.com/neekrasov/kvdb/internal/database/storage"
	mocks "github.com/neekrasov/kvdb/internal/mocks/database"
	"github.com/neekrasov/kvdb/pkg/gob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUsersStorage(t *testing.T) {
	mockStorage := mocks.NewStorage(t)
	usersStorage := identity.NewUsersStorage(mockStorage)

	t.Run("Test Create - success", func(t *testing.T) {
		username := "testUser"
		password := "testPass"
		key := storage.MakeKey(models.SystemUserNameSpace, username)

		mockStorage.On("Get", key).Return("", storage.ErrKeyNotFound).Once()
		mockStorage.On("Set", key, mock.Anything).Return(nil).Once()

		_, err := usersStorage.Create(username, password)
		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test Create - already exists", func(t *testing.T) {
		username := "existingUser"
		password := "testPass"
		key := storage.MakeKey(models.SystemUserNameSpace, username)

		mockStorage.On("Get", key).Return("{}", nil).Once()

		_, err := usersStorage.Create(username, password)
		assert.Equal(t, identity.ErrUserAlreadyExists, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test Authenticate - success", func(t *testing.T) {
		username := "authUser"
		password := "authPass"
		key := storage.MakeKey(models.SystemUserNameSpace, username)

		user := models.User{Username: username, Password: password}
		userBytes, _ := gob.Encode(user)

		mockStorage.On("Get", key).Return(string(userBytes), nil).Once()

		retrievedUser, err := usersStorage.Authenticate(username, password)
		assert.NoError(t, err)
		assert.Equal(t, username, retrievedUser.Username)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test Authenticate - wrong password", func(t *testing.T) {
		username := "authUser"
		password := "authPass"
		key := storage.MakeKey(models.SystemUserNameSpace, username)

		user := models.User{Username: username, Password: "wrongPass"}
		userBytes, _ := gob.Encode(user)

		mockStorage.On("Get", key).Return(string(userBytes), nil).Once()

		_, err := usersStorage.Authenticate(username, password)
		assert.Equal(t, identity.ErrAuthenticationFailed, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test Authenticate - user not found", func(t *testing.T) {
		username := "nonexistentUser"
		password := "testPass"
		key := storage.MakeKey(models.SystemUserNameSpace, username)

		mockStorage.On("Get", key).Return("", storage.ErrKeyNotFound).Once()

		_, err := usersStorage.Authenticate(username, password)
		assert.Equal(t, identity.ErrAuthenticationFailed, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test AssignRole - success", func(t *testing.T) {
		username := "roleUser"
		role := "admin"
		userKey := storage.MakeKey(models.SystemUserNameSpace, username)
		roleKey := storage.MakeKey(models.SystemRoleNameSpace, role)

		user := models.User{Username: username, Roles: []string{}}
		userBytes, _ := gob.Encode(user)

		mockStorage.On("Get", userKey).Return(string(userBytes), nil).Once()
		mockStorage.On("Get", roleKey).Return("{}", nil).Once()
		mockStorage.On("Set", userKey, mock.Anything).Return(nil).Once()

		err := usersStorage.AssignRole(username, role)
		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test AssignRole - user not found", func(t *testing.T) {
		username := "nonexistentUser"
		role := "admin"
		userKey := storage.MakeKey(models.SystemUserNameSpace, username)

		mockStorage.On("Get", userKey).Return("", storage.ErrKeyNotFound).Once()

		err := usersStorage.AssignRole(username, role)
		assert.Equal(t, identity.ErrUserNotFound, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test AssignRole - role not found", func(t *testing.T) {
		username := "roleUser"
		role := "nonexistentRole"
		userKey := storage.MakeKey(models.SystemUserNameSpace, username)
		roleKey := storage.MakeKey(models.SystemRoleNameSpace, role)

		user := models.User{Username: username, Roles: []string{}}
		userBytes, _ := gob.Encode(user)

		mockStorage.On("Get", userKey).Return(string(userBytes), nil).Once()
		mockStorage.On("Get", roleKey).Return("", storage.ErrKeyNotFound).Once()

		err := usersStorage.AssignRole(username, role)
		assert.Equal(t, identity.ErrRoleNotFound, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test Delete - success", func(t *testing.T) {
		username := "deleteUser"
		key := storage.MakeKey(models.SystemUserNameSpace, username)

		mockStorage.On("Get", key).Return("{}", nil).Once()
		mockStorage.On("Del", key).Return(nil).Once()

		err := usersStorage.Delete(username)
		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test Delete - not found", func(t *testing.T) {
		username := "nonexistentUser"
		key := storage.MakeKey(models.SystemUserNameSpace, username)

		mockStorage.On("Get", key).Return("", storage.ErrKeyNotFound).Once()

		err := usersStorage.Delete(username)
		assert.Equal(t, identity.ErrUserNotFound, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test SaveRaw - success", func(t *testing.T) {
		user := models.User{Username: "testUser", Password: "testPass"}
		key := storage.MakeKey(models.SystemUserNameSpace, user.Username)

		mockStorage.On("Set", key, mock.Anything).Return(nil).Once()

		err := usersStorage.SaveRaw(&user)
		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test Get - success", func(t *testing.T) {
		username := "testUser"
		key := storage.MakeKey(models.SystemUserNameSpace, username)
		user := models.User{Username: username, Password: "testPass"}
		userBytes, _ := gob.Encode(user)

		mockStorage.On("Get", key).Return(string(userBytes), nil).Once()

		retrievedUser, err := usersStorage.Get(username)
		assert.NoError(t, err)
		assert.Equal(t, username, retrievedUser.Username)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test Get - not found", func(t *testing.T) {
		username := "nonexistentUser"
		key := storage.MakeKey(models.SystemUserNameSpace, username)

		mockStorage.On("Get", key).Return("", storage.ErrKeyNotFound).Once()

		_, err := usersStorage.Get(username)
		assert.Equal(t, identity.ErrUserNotFound, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test Get - JSON unmarshal error", func(t *testing.T) {
		username := "testUser"
		key := storage.MakeKey(models.SystemUserNameSpace, username)

		mockStorage.On("Get", key).Return("invalid json", nil).Once()

		_, err := usersStorage.Get(username)
		assert.Error(t, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test Append - success", func(t *testing.T) {
		username := "newUser"
		key := models.SystemUsersKey

		listBytes, _ := gob.Encode([]string{})

		mockStorage.On("Get", key).Return(string(listBytes), nil).Once()
		mockStorage.On("Set", key, mock.Anything).Return(nil).Once()

		users, err := usersStorage.Append(username)
		assert.NoError(t, err)
		assert.Contains(t, users, username)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test ListUsernames - success", func(t *testing.T) {
		key := models.SystemUsersKey
		users := []string{"user1", "user2"}
		usersBytes, _ := gob.Encode(users)

		mockStorage.On("Get", key).Return(string(usersBytes), nil).Once()

		retrievedUsers, err := usersStorage.ListUsernames()
		assert.NoError(t, err)
		assert.Equal(t, users, retrievedUsers)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test ListUsernames - empty list", func(t *testing.T) {
		key := models.SystemUsersKey

		mockStorage.On("Get", key).Return("", storage.ErrKeyNotFound).Once()

		users, err := usersStorage.ListUsernames()
		assert.Error(t, err, identity.ErrEmptyUsers)
		assert.Empty(t, users)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test ListUsernames - decode error", func(t *testing.T) {
		key := models.SystemUsersKey

		mockStorage.On("Get", key).Return("invalid json", nil).Once()

		_, err := usersStorage.ListUsernames()
		assert.Error(t, err)
		mockStorage.AssertExpectations(t)
	})
}
