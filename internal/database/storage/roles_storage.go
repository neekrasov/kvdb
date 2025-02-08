package storage

import (
	"github.com/neekrasov/kvdb/internal/database/storage/models"
	"github.com/neekrasov/kvdb/pkg/gob"
)

// RolesStorage - A struct that manages role-related operations, such as creating, deleting, and listing roles.
type RolesStorage struct {
	engine Engine
}

// NewRolesStorage - Initializes and returns a new RolesStorage instance with the provided storage engine.
func NewRolesStorage(
	engine Engine,
) *RolesStorage {
	return &RolesStorage{
		engine: engine,
	}
}

// Get - Retrieves a role by its name.
func (s *RolesStorage) Get(name string) (*models.Role, error) {
	key := MakeKey(models.SystemRoleNameSpace, name)
	roleBytes, exists := s.engine.Get(key)
	if !exists {
		return nil, models.ErrRoleNotFound
	}

	var role models.Role
	if err := gob.Decode([]byte(roleBytes), &role); err != nil {
		return nil, err
	}

	return &role, nil
}

// Delete - Deletes a role by its name.
func (s *RolesStorage) Delete(name string) error {
	key := MakeKey(models.SystemRoleNameSpace, name)
	if _, exists := s.engine.Get(key); !exists {
		return models.ErrRoleNotFound
	}

	return s.engine.Del(key)
}

// Append - Adds a new role to the list of all roles in the system.
func (s *RolesStorage) Append(role string) ([]string, error) {
	roles, err := s.List()
	if err != nil {
		return nil, err
	}
	roles = append(roles, role)

	rolesBytes, err := gob.Encode(roles)
	if err != nil {
		return roles, err
	}
	s.engine.Set(models.SystemRolesKey, string(rolesBytes))

	return roles, nil
}

// List - Retrieves a list of all roles in the system.
func (s *RolesStorage) List() ([]string, error) {
	var roles []string
	rolesString, exists := s.engine.Get(models.SystemRolesKey)
	if !exists {
		return roles, nil
	}

	if err := gob.Decode([]byte(rolesString), &roles); err != nil {
		return nil, err
	}

	return roles, nil
}

// Save - Saves a role to the storage.
func (s *RolesStorage) Save(role *models.Role) error {
	key := MakeKey(models.SystemRoleNameSpace, role.Name)
	if _, exists := s.engine.Get(key); exists {
		return models.ErrRoleAlreadyExists
	}

	roleBytes, err := gob.Encode(role)
	if err != nil {
		return err
	}
	s.engine.Set(key, string(roleBytes))

	return nil
}
