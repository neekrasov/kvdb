package storage

import (
	"github.com/neekrasov/kvdb/internal/database/models"
	"github.com/neekrasov/kvdb/pkg/gob"
)

// UsersStorage - A struct that manages user-related operations,
// such as authentication, user creation, and role assignment.
type UsersStorage struct {
	engine Engine
}

// NewUsersStorage - Initializes and returns a new UsersStorage instance with the provided storage engine.
func NewUsersStorage(engine Engine) *UsersStorage {
	return &UsersStorage{
		engine: engine,
	}
}

// Authenticate - Authenticates a user by verifying their username and password.
func (s *UsersStorage) Authenticate(username, password string) (*models.User, error) {
	userKey := MakeKey(models.SystemUserNameSpace, username)
	userString, exists := s.engine.Get(userKey)
	if !exists {
		return nil, models.ErrAuthenticationFailed
	}

	var user models.User
	if err := gob.Decode([]byte(userString), &user); err != nil {
		return nil, err
	}

	if user.Password != password {
		return nil, models.ErrAuthenticationFailed
	}

	return &user, nil
}

// AssignRole - Assigns a role to a user.
func (s *UsersStorage) AssignRole(username string, role string) error {
	userKey := MakeKey(models.SystemUserNameSpace, username)
	userString, exists := s.engine.Get(userKey)
	if !exists {
		return models.ErrUserNotFound
	}

	roleKey := MakeKey(models.SystemRoleNameSpace, role)
	if _, exists := s.engine.Get(roleKey); !exists {
		return models.ErrRoleNotFound
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
	s.engine.Set(userKey, string(userBytesUpdated))

	return nil
}

// Create - Creates a new user with the specified username and password.
func (s *UsersStorage) Create(username, password string) (*models.User, error) {
	key := MakeKey(models.SystemUserNameSpace, username)
	if _, exists := s.engine.Get(key); exists {
		return nil, models.ErrUserAlreadyExists
	}

	user := models.User{
		Username: username,
		Password: password,
		Roles:    []string{models.DefaultRoleName},
		Cur:      models.DefaultRole,
	}

	userBytes, err := gob.Encode(user)
	if err != nil {
		return nil, err
	}
	s.engine.Set(key, string(userBytes))

	return &user, nil
}

// SaveRaw - Saves a user object directly to the storage.
func (s *UsersStorage) SaveRaw(user *models.User) error {
	key := MakeKey(models.SystemUserNameSpace, user.Username)
	userBytes, err := gob.Encode(user)
	if err != nil {
		return err
	}
	s.engine.Set(key, string(userBytes))

	return nil
}

// Get - Retrieves a user by their username.
func (s *UsersStorage) Get(username string) (*models.User, error) {

	key := MakeKey(models.SystemUserNameSpace, username)
	userString, exists := s.engine.Get(key)
	if !exists {
		return nil, models.ErrUserNotFound
	}

	var user models.User
	if err := gob.Decode([]byte(userString), &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// Delete - Deletes a user by their username.
func (s *UsersStorage) Delete(username string) error {
	key := MakeKey(models.SystemUserNameSpace, username)
	if _, exists := s.engine.Get(key); !exists {
		return models.ErrUserNotFound
	}

	return s.engine.Del(key)
}

// Append - Adds a new username to the list of all users in the system.
func (s *UsersStorage) Append(user string) ([]string, error) {
	users, err := s.ListUsernames()
	if err != nil {
		return nil, err
	}
	users = append(users, user)

	usersBytes, err := gob.Encode(users)
	if err != nil {
		return users, err
	}
	s.engine.Set(models.SystemUsersKey, string(usersBytes))

	return users, nil
}

// ListUsernames - Retrieves a list of all usernames in the system.
func (s *UsersStorage) ListUsernames() ([]string, error) {
	var users []string
	userString, exists := s.engine.Get(models.SystemUsersKey)
	if !exists {
		return users, nil
	}

	if err := gob.Decode([]byte(userString), &users); err != nil {
		return nil, err
	}

	return users, nil
}
