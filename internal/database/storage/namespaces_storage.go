package storage

import (
	"github.com/neekrasov/kvdb/internal/database/storage/models"
	"github.com/neekrasov/kvdb/pkg/gob"
)

// NamespaceStorage - A struct that manages namespace-related operations,
// such as creating, deleting, and listing namespaces.
type NamespaceStorage struct {
	engine Engine
}

// NewNamespaceStorage - Initializes and returns a new NamespaceStorage instance with the provided storage engine.
func NewNamespaceStorage(
	engine Engine,
) *NamespaceStorage {
	return &NamespaceStorage{
		engine: engine,
	}
}

// Exists - Checks if a namespace exists in the storage. Returns true if it exists, otherwise false.
func (s *NamespaceStorage) Exists(namespace string) bool {
	key := MakeKey(models.SystemNamespaceNameSpace, namespace)
	_, exists := s.engine.Get(key)

	return exists
}

// Save - Saves a new namespace to the storage.
func (s *NamespaceStorage) Save(namespace string) error {
	key := MakeKey(models.SystemNamespaceNameSpace, namespace)
	if _, exists := s.engine.Get(key); exists {
		return models.ErrNamespaceAlreadyExists
	}
	s.engine.Set(key, "{}")

	return nil
}

// Delete - Deletes a namespace from the storage.
func (s *NamespaceStorage) Delete(namespace string) error {
	key := MakeKey(models.SystemNamespaceNameSpace, namespace)
	if _, exists := s.engine.Get(key); !exists {
		return models.ErrNamespaceNotFound
	}

	return s.engine.Del(key)
}

// Append - Adds a new namespace to the list of all namespaces in the system.
func (s *NamespaceStorage) Append(namespace string) ([]string, error) {
	namespaces, err := s.List()
	if err != nil {
		return nil, err
	}
	namespaces = append(namespaces, namespace)

	usersBytes, err := gob.Encode(namespaces)
	if err != nil {
		return namespaces, err
	}
	s.engine.Set(models.SystemNamespacesKey, string(usersBytes))

	return namespaces, nil
}

// List - Retrieves a list of all namespaces in the system.
func (s *NamespaceStorage) List() ([]string, error) {
	var namespaces []string
	namespacesString, exists := s.engine.Get(models.SystemNamespacesKey)
	if !exists {
		return namespaces, nil
	}

	if err := gob.Decode([]byte(namespacesString), &namespaces); err != nil {
		return nil, err
	}

	return namespaces, nil
}
