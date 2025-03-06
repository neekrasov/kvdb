package identity

import (
	"errors"

	"github.com/neekrasov/kvdb/internal/database/identity/models"
	"github.com/neekrasov/kvdb/internal/database/storage"
	"github.com/neekrasov/kvdb/pkg/gob"
)

var (
	ErrRoleAlreadyExists = errors.New("role already exists")
	ErrRoleNotFound      = errors.New("role not found")
	ErrEmptyRoles        = errors.New("empty roles")
)

// RolesStorage - struct that manages role-related operations, such as creating, deleting, and listing roles.
type RolesStorage struct {
	storage Storage
}

// NewRolesStorage - initializes and returns a new RolesStorage instance with the provided storage engine.
func NewRolesStorage(storage Storage) *RolesStorage {
	return &RolesStorage{storage: storage}
}

// Get - retrieves a role by its name.
func (s *RolesStorage) Get(name string) (*models.Role, error) {
	key := storage.MakeKey(models.SystemRoleNameSpace, name)
	roleBytes, err := s.storage.Get(key)
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			return nil, ErrRoleNotFound
		}

		return nil, err
	}

	var role models.Role
	if err := gob.Decode([]byte(roleBytes), &role); err != nil {
		return nil, err
	}

	return &role, nil
}

// Delete - deletes a role by its name.
func (s *RolesStorage) Delete(name string) error {
	key := storage.MakeKey(models.SystemRoleNameSpace, name)
	if _, err := s.storage.Get(key); err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			return ErrRoleNotFound
		}

		return err
	}

	return s.storage.Del(key)
}

// Append - adds a new role to the list of all roles in the system.
func (s *RolesStorage) Append(role string) ([]string, error) {
	roles, err := s.List()
	if err != nil && !errors.Is(err, ErrRoleNotFound) {
		return nil, err
	}
	roles = append(roles, role)

	rolesBytes, err := gob.Encode(roles)
	if err != nil {
		return roles, err
	}

	err = s.storage.Set(models.SystemRolesKey, string(rolesBytes))
	if err != nil {
		return nil, err
	}

	return roles, nil
}

// List - retrieves a list of all roles in the system.
func (s *RolesStorage) List() ([]string, error) {
	rolesString, err := s.storage.Get(models.SystemRolesKey)
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			return nil, ErrRoleNotFound
		}

		return nil, err
	}

	var roles []string
	if err := gob.Decode([]byte(rolesString), &roles); err != nil {
		return nil, err
	}

	return roles, nil
}

// Save - saves a role to the storage.
func (s *RolesStorage) Save(role *models.Role) error {
	key := storage.MakeKey(models.SystemRoleNameSpace, role.Name)
	roleString, err := s.storage.Get(key)
	if err != nil && !errors.Is(err, storage.ErrKeyNotFound) {
		return err
	}

	if roleString != "" {
		return ErrRoleAlreadyExists
	}

	roleBytes, err := gob.Encode(role)
	if err != nil {
		return err
	}

	return s.storage.Set(key, string(roleBytes))
}
