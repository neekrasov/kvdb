package storage

import (
	"testing"

	"github.com/neekrasov/kvdb/internal/database/models"
	mocks "github.com/neekrasov/kvdb/internal/mocks/storage"
	"github.com/neekrasov/kvdb/pkg/gob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUsersStorage(t *testing.T) {
	mockEngine := mocks.NewEngine(t)
	usersStorage := NewUsersStorage(mockEngine)

	t.Run("Test Create - success", func(t *testing.T) {
		username := "testUser"
		password := "testPass"
		key := MakeKey(models.SystemUserNameSpace, username)

		mockEngine.On("Get", key).Return("", false).Once()
		mockEngine.On("Set", key, mock.Anything).Return().Once()

		_, err := usersStorage.Create(username, password)
		assert.NoError(t, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test Create - already exists", func(t *testing.T) {
		username := "existingUser"
		password := "testPass"
		key := MakeKey(models.SystemUserNameSpace, username)

		mockEngine.On("Get", key).Return("{}", true).Once()

		_, err := usersStorage.Create(username, password)
		assert.Equal(t, models.ErrUserAlreadyExists, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test Authenticate - success", func(t *testing.T) {
		username := "authUser"
		password := "authPass"
		key := MakeKey(models.SystemUserNameSpace, username)

		user := models.User{Username: username, Password: password}
		userBytes, _ := gob.Encode(user)

		mockEngine.On("Get", key).Return(string(userBytes), true).Once()

		retrievedUser, err := usersStorage.Authenticate(username, password)
		assert.NoError(t, err)
		assert.Equal(t, username, retrievedUser.Username)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test Authenticate - wrong password", func(t *testing.T) {
		username := "authUser"
		password := "authPass"
		key := MakeKey(models.SystemUserNameSpace, username)

		user := models.User{Username: username, Password: "wrongPass"}
		userBytes, _ := gob.Encode(user)

		mockEngine.On("Get", key).Return(string(userBytes), true).Once()

		_, err := usersStorage.Authenticate(username, password)
		assert.Equal(t, models.ErrAuthenticationFailed, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test Authenticate - user not found", func(t *testing.T) {
		username := "nonexistentUser"
		password := "testPass"
		key := MakeKey(models.SystemUserNameSpace, username)

		mockEngine.On("Get", key).Return("", false).Once()

		_, err := usersStorage.Authenticate(username, password)
		assert.Equal(t, models.ErrAuthenticationFailed, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test AssignRole - success", func(t *testing.T) {
		username := "roleUser"
		role := "admin"
		userKey := MakeKey(models.SystemUserNameSpace, username)
		roleKey := MakeKey(models.SystemRoleNameSpace, role)

		user := models.User{Username: username, Roles: []string{}}
		userBytes, _ := gob.Encode(user)

		mockEngine.On("Get", userKey).Return(string(userBytes), true).Once()
		mockEngine.On("Get", roleKey).Return("{}", true).Once()
		mockEngine.On("Set", userKey, mock.Anything).Return().Once()

		err := usersStorage.AssignRole(username, role)
		assert.NoError(t, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test AssignRole - user not found", func(t *testing.T) {
		username := "nonexistentUser"
		role := "admin"
		userKey := MakeKey(models.SystemUserNameSpace, username)

		mockEngine.On("Get", userKey).Return("", false).Once()

		err := usersStorage.AssignRole(username, role)
		assert.Equal(t, models.ErrUserNotFound, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test AssignRole - role not found", func(t *testing.T) {
		username := "roleUser"
		role := "nonexistentRole"
		userKey := MakeKey(models.SystemUserNameSpace, username)
		roleKey := MakeKey(models.SystemRoleNameSpace, role)

		user := models.User{Username: username, Roles: []string{}}
		userBytes, _ := gob.Encode(user)

		mockEngine.On("Get", userKey).Return(string(userBytes), true).Once()
		mockEngine.On("Get", roleKey).Return("", false).Once()

		err := usersStorage.AssignRole(username, role)
		assert.Equal(t, models.ErrRoleNotFound, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test Delete - success", func(t *testing.T) {
		username := "deleteUser"
		key := MakeKey(models.SystemUserNameSpace, username)

		mockEngine.On("Get", key).Return("{}", true).Once()
		mockEngine.On("Del", key).Return(nil).Once()

		err := usersStorage.Delete(username)
		assert.NoError(t, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test Delete - not found", func(t *testing.T) {
		username := "nonexistentUser"
		key := MakeKey(models.SystemUserNameSpace, username)

		mockEngine.On("Get", key).Return("", false).Once()

		err := usersStorage.Delete(username)
		assert.Equal(t, models.ErrUserNotFound, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test SaveRaw - success", func(t *testing.T) {
		user := models.User{Username: "testUser", Password: "testPass"}
		key := MakeKey(models.SystemUserNameSpace, user.Username)

		mockEngine.On("Set", key, mock.Anything).Return().Once()

		err := usersStorage.SaveRaw(&user)
		assert.NoError(t, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test Get - success", func(t *testing.T) {
		username := "testUser"
		key := MakeKey(models.SystemUserNameSpace, username)
		user := models.User{Username: username, Password: "testPass"}
		userBytes, _ := gob.Encode(user)

		mockEngine.On("Get", key).Return(string(userBytes), true).Once()

		retrievedUser, err := usersStorage.Get(username)
		assert.NoError(t, err)
		assert.Equal(t, username, retrievedUser.Username)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test Get - not found", func(t *testing.T) {
		username := "nonexistentUser"
		key := MakeKey(models.SystemUserNameSpace, username)

		mockEngine.On("Get", key).Return("", false).Once()

		_, err := usersStorage.Get(username)
		assert.Equal(t, models.ErrUserNotFound, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test Get - JSON unmarshal error", func(t *testing.T) {
		username := "testUser"
		key := MakeKey(models.SystemUserNameSpace, username)

		mockEngine.On("Get", key).Return("invalid json", true).Once()

		_, err := usersStorage.Get(username)
		assert.Error(t, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test Append - success", func(t *testing.T) {
		username := "newUser"
		key := models.SystemUsersKey

		listBytes, _ := gob.Encode([]string{})

		mockEngine.On("Get", key).Return(string(listBytes), true).Once()
		mockEngine.On("Set", key, mock.Anything).Return().Once()

		users, err := usersStorage.Append(username)
		assert.NoError(t, err)
		assert.Contains(t, users, username)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test ListUsernames - success", func(t *testing.T) {
		key := models.SystemUsersKey
		users := []string{"user1", "user2"}
		usersBytes, _ := gob.Encode(users)

		mockEngine.On("Get", key).Return(string(usersBytes), true).Once()

		retrievedUsers, err := usersStorage.ListUsernames()
		assert.NoError(t, err)
		assert.Equal(t, users, retrievedUsers)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test ListUsernames - empty list", func(t *testing.T) {
		key := models.SystemUsersKey

		mockEngine.On("Get", key).Return("", false).Once()

		users, err := usersStorage.ListUsernames()
		assert.NoError(t, err)
		assert.Empty(t, users)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test ListUsernames - decode error", func(t *testing.T) {
		key := models.SystemUsersKey

		mockEngine.On("Get", key).Return("invalid json", true).Once()

		_, err := usersStorage.ListUsernames()
		assert.Error(t, err)
		mockEngine.AssertExpectations(t)
	})
}
