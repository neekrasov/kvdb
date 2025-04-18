package identity

import (
	"context"
	"errors"

	"github.com/neekrasov/kvdb/internal/database/identity/models"
	"github.com/neekrasov/kvdb/internal/database/storage"
	"github.com/neekrasov/kvdb/pkg/gob"
	"golang.org/x/crypto/bcrypt"
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
func (s *UsersStorage) Authenticate(ctx context.Context, username, password string) (*models.User, error) {
	userKey := storage.MakeKey(models.SystemUserNameSpace, username)
	userString, err := s.storage.Get(ctx, userKey)
	if err != nil {
		return nil, ErrAuthenticationFailed
	}

	var user models.User
	if err := gob.Decode([]byte(userString), &user); err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrAuthenticationFailed
	}

	return &user, nil
}

// AssignRole - assigns a role to a user.
func (s *UsersStorage) AssignRole(ctx context.Context, username string, role string) error {
	userKey := storage.MakeKey(models.SystemUserNameSpace, username)
	userString, err := s.storage.Get(ctx, userKey)
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			return ErrUserNotFound
		}

		return err
	}

	roleKey := storage.MakeKey(models.SystemRoleNameSpace, role)
	if _, err = s.storage.Get(ctx, roleKey); err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			return ErrRoleNotFound
		}

		return err
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

	err = s.storage.Set(ctx, userKey, string(userBytesUpdated))
	if err != nil {
		return err
	}

	return nil
}

// AssignRole - assigns a role to a user.
func (s *UsersStorage) DivestRole(ctx context.Context, username string, role string) error {
	userKey := storage.MakeKey(models.SystemUserNameSpace, username)
	userString, err := s.storage.Get(ctx, userKey)
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			return ErrUserNotFound
		}

		return err
	}

	roleKey := storage.MakeKey(models.SystemRoleNameSpace, role)
	if _, err = s.storage.Get(ctx, roleKey); err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			return ErrRoleNotFound
		}

		return err
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

	if index == -1 {
		return nil
	}

	user.Roles[index] = user.Roles[len(user.Roles)-1]
	user.Roles = user.Roles[:len(user.Roles)-1]
	userBytesUpdated, err := gob.Encode(user)
	if err != nil {
		return err
	}

	err = s.storage.Set(ctx, userKey, string(userBytesUpdated))
	if err != nil {
		return err
	}

	return nil
}

// Create - creates a new user with the specified username and password.
func (s *UsersStorage) Create(ctx context.Context, username, password string) (*models.User, error) {
	key := storage.MakeKey(models.SystemUserNameSpace, username)
	if _, err := s.storage.Get(ctx, key); err == nil {
		return nil, ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := models.User{
		Username:   username,
		Password:   string(hashedPassword),
		Roles:      []string{models.DefaultRoleName},
		ActiveRole: models.DefaultRole,
	}

	userBytes, err := gob.Encode(user)
	if err != nil {
		return nil, err
	}

	err = s.storage.Set(ctx, key, string(userBytes))
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// SaveRaw - saves a user object directly to the storage.
func (s *UsersStorage) SaveRaw(ctx context.Context, user *models.User) error {
	key := storage.MakeKey(models.SystemUserNameSpace, user.Username)
	if _, err := s.storage.Get(ctx, key); err == nil {
		return ErrUserAlreadyExists
	}

	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hashedPassword)
	}

	userBytes, err := gob.Encode(user)
	if err != nil {
		return err
	}

	return s.storage.Set(ctx, key, string(userBytes))
}

// Get - retrieves a user by their username.
func (s *UsersStorage) Get(ctx context.Context, username string) (*models.User, error) {
	key := storage.MakeKey(models.SystemUserNameSpace, username)
	userString, err := s.storage.Get(ctx, key)
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
func (s *UsersStorage) Delete(ctx context.Context, username string) error {
	key := storage.MakeKey(models.SystemUserNameSpace, username)
	if _, err := s.storage.Get(ctx, key); err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			return ErrUserNotFound
		}

		return err
	}

	return s.storage.Del(ctx, key)
}

// Append - adds a new username to the list of all users in the system.
func (s *UsersStorage) Append(ctx context.Context, user string) ([]string, error) {
	users, err := s.ListUsernames(ctx)
	if err != nil && !errors.Is(err, ErrEmptyUsers) {
		return nil, err
	}
	users = append(users, user)

	usersBytes, err := gob.Encode(users)
	if err != nil {
		return users, err
	}

	err = s.storage.Set(ctx, models.SystemUsersKey, string(usersBytes))
	if err != nil {
		return nil, err
	}

	return users, nil
}

// Remove - remove username from the list of all users in the system.
func (s *UsersStorage) Remove(ctx context.Context, user string) ([]string, error) {
	users, err := s.ListUsernames(ctx)
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

	if index == -1 {
		return users, nil
	}

	users = append(users[:index], users[index+1:]...)
	usersBytes, err := gob.Encode(users)
	if err != nil {
		return users, err
	}

	err = s.storage.Set(ctx, models.SystemUsersKey, string(usersBytes))
	if err != nil {
		return nil, err
	}

	return users, nil
}

// ListUsernames - retrieves a list of all usernames in the system.
func (s *UsersStorage) ListUsernames(ctx context.Context) ([]string, error) {
	usersString, err := s.storage.Get(ctx, models.SystemUsersKey)
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
