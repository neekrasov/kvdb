package identity

import (
	"errors"

	"github.com/neekrasov/kvdb/internal/database/identity/models"
	"github.com/neekrasov/kvdb/internal/database/storage"
	"github.com/neekrasov/kvdb/pkg/gob"
)

var (
	ErrUserAlreadyExists    = errors.New("user already exists")
	ErrUserNotFound         = errors.New("user not found")
	ErrEmptyUsers           = errors.New("empty users")
	ErrAuthenticationFailed = errors.New("authentication failed")
)

// UsersStorage - a struct that manages user-related operations,
// such as authentication, user creation, and role assignment.
type UsersStorage struct {
	storage Storage
}

// NewUsersStorage - initializes and returns a new UsersStorage instance with the provided storage engine.
func NewUsersStorage(storage Storage) *UsersStorage {
	return &UsersStorage{storage: storage}
}

// Authenticate - authenticates a user by verifying their username and password.
func (s *UsersStorage) Authenticate(username, password string) (*models.User, error) {
	userKey := storage.MakeKey(models.SystemUserNameSpace, username)
	userString, err := s.storage.Get(userKey)
	if err != nil {
		return nil, ErrAuthenticationFailed
	}

	var user models.User
	if err := gob.Decode([]byte(userString), &user); err != nil {
		return nil, err
	}

	if user.Password != password {
		return nil, ErrAuthenticationFailed
	}

	return &user, nil
}

// AssignRole - assigns a role to a user.
func (s *UsersStorage) AssignRole(username string, role string) error {
	userKey := storage.MakeKey(models.SystemUserNameSpace, username)
	userString, err := s.storage.Get(userKey)
	if err != nil && errors.Is(err, storage.ErrKeyNotFound) {
		return ErrUserNotFound
	}

	roleKey := storage.MakeKey(models.SystemRoleNameSpace, role)
	_, err = s.storage.Get(roleKey)
	if err != nil && errors.Is(err, storage.ErrKeyNotFound) {
		return ErrRoleNotFound
	}

	var user models.User
	if err := gob.Decode([]byte(userString), &user); err != nil {
		return err
	}

	user.Roles = append(user.Roles, role)
	userBytesUpdated, err := gob.Encode(user)
	if err != nil {
		return err
	}

	err = s.storage.Set(userKey, string(userBytesUpdated))
	if err != nil {
		return err
	}

	return nil
}

// AssignRole - assigns a role to a user.
func (s *UsersStorage) DivestRole(username string, role string) error {
	userKey := storage.MakeKey(models.SystemUserNameSpace, username)
	userString, err := s.storage.Get(userKey)
	if err != nil && errors.Is(err, storage.ErrKeyNotFound) {
		return ErrUserNotFound
	}

	roleKey := storage.MakeKey(models.SystemRoleNameSpace, role)
	_, err = s.storage.Get(roleKey)
	if err != nil && errors.Is(err, storage.ErrKeyNotFound) {
		return ErrRoleNotFound
	}

	var user models.User
	if err := gob.Decode([]byte(userString), &user); err != nil {
		return err
	}

	index := -1
	for i, roleName := range user.Roles {
		if roleName == role {
			index = i
			break
		}
	}

	user.Roles[index] = user.Roles[len(user.Roles)-1]
	user.Roles = user.Roles[:len(user.Roles)-1]
	userBytesUpdated, err := gob.Encode(user)
	if err != nil {
		return err
	}

	err = s.storage.Set(userKey, string(userBytesUpdated))
	if err != nil {
		return err
	}

	return nil
}

// Create - creates a new user with the specified username and password.
func (s *UsersStorage) Create(username, password string) (*models.User, error) {
	key := storage.MakeKey(models.SystemUserNameSpace, username)
	if _, err := s.storage.Get(key); err == nil {
		return nil, ErrUserAlreadyExists
	}

	user := models.User{
		Username:   username,
		Password:   password,
		Roles:      []string{models.DefaultRoleName},
		ActiveRole: models.DefaultRole,
	}

	userBytes, err := gob.Encode(user)
	if err != nil {
		return nil, err
	}

	err = s.storage.Set(key, string(userBytes))
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// SaveRaw - saves a user object directly to the storage.
func (s *UsersStorage) SaveRaw(user *models.User) error {
	key := storage.MakeKey(models.SystemUserNameSpace, user.Username)
	if _, err := s.storage.Get(key); err == nil {
		return ErrUserAlreadyExists
	}

	userBytes, err := gob.Encode(user)
	if err != nil {
		return err
	}

	return s.storage.Set(key, string(userBytes))
}

// Get - retrieves a user by their username.
func (s *UsersStorage) Get(username string) (*models.User, error) {
	key := storage.MakeKey(models.SystemUserNameSpace, username)
	userString, err := s.storage.Get(key)
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	var user models.User
	if err := gob.Decode([]byte(userString), &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// Delete - deletes a user by their username.
func (s *UsersStorage) Delete(username string) error {
	key := storage.MakeKey(models.SystemUserNameSpace, username)
	if _, err := s.storage.Get(key); err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			return ErrUserNotFound
		}

		return err
	}

	return s.storage.Del(key)
}

// Append - adds a new username to the list of all users in the system.
func (s *UsersStorage) Append(user string) ([]string, error) {
	users, err := s.ListUsernames()
	if err != nil && !errors.Is(err, ErrEmptyUsers) {
		return nil, err
	}
	users = append(users, user)

	usersBytes, err := gob.Encode(users)
	if err != nil {
		return users, err
	}

	err = s.storage.Set(models.SystemUsersKey, string(usersBytes))
	if err != nil {
		return nil, err
	}

	return users, nil
}

// Remove - remove username from the list of all users in the system.
func (s *UsersStorage) Remove(user string) ([]string, error) {
	users, err := s.ListUsernames()
	if err != nil && !errors.Is(err, ErrEmptyUsers) {
		return nil, err
	}

	if len(users) == 0 {
		return nil, nil
	}

	index := -1
	for i, u := range users {
		if u == user {
			index = i
			break
		}
	}
	users[index] = users[len(users)-1]

	usersBytes, err := gob.Encode(users[:len(users)-1])
	if err != nil {
		return users, err
	}

	err = s.storage.Set(models.SystemUsersKey, string(usersBytes))
	if err != nil {
		return nil, err
	}

	return users, nil
}

// ListUsernames - retrieves a list of all usernames in the system.
func (s *UsersStorage) ListUsernames() ([]string, error) {
	usersString, err := s.storage.Get(models.SystemUsersKey)
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			return nil, ErrEmptyUsers
		}

		return nil, err
	}

	var users []string
	if err := gob.Decode([]byte(usersString), &users); err != nil {
		return nil, err
	}

	return users, nil
}
