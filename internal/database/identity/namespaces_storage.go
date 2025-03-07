package identity

import (
	"context"
	"errors"
	"io"

	"github.com/neekrasov/kvdb/internal/database/identity/models"
	"github.com/neekrasov/kvdb/internal/database/storage"

	"github.com/neekrasov/kvdb/pkg/gob"
)

var (
	ErrNamespaceAlreadyExists = errors.New("namespace already exists")
	ErrNamespaceNotFound      = errors.New("namespace not found")
	ErrEmptyNamespaces        = errors.New("empty namespaces")
)

// Storage - interface for storing, retrieving, and deleting key-value pairs.
type Storage interface {
	// Set - stores a value for a given key.
	Set(ctx context.Context, key, value string) error
	// Get - retrieves the value associated with a given key.
	Get(ctx context.Context, key string) (string, error)
	// Del - removes a key and its value from the storage.
	Del(ctx context.Context, key string) error
}

// NamespaceStorage - struct that manages namespace-related operations,
// such as creating, deleting, and listing namespaces.
type NamespaceStorage struct {
	storage Storage
}

// NewNamespaceStorage - initializes and returns a new NamespaceStorage instance with the provided storage engine.
func NewNamespaceStorage(storage Storage) *NamespaceStorage {
	return &NamespaceStorage{storage: storage}
}

// Exists - checks if a namespace exists in the storage. Returns true if it exists, otherwise false.
func (s *NamespaceStorage) Exists(ctx context.Context, namespace string) bool {
	key := storage.MakeKey(models.SystemNamespaceNameSpace, namespace)
	if _, err := s.storage.Get(ctx, key); err != nil {
		return false
	}

	return true
}

// Save - saves a new namespace to the storage.
func (s *NamespaceStorage) Save(ctx context.Context, namespace string) error {
	key := storage.MakeKey(models.SystemNamespaceNameSpace, namespace)
	if _, err := s.storage.Get(ctx, key); err == nil {
		return ErrNamespaceAlreadyExists
	}

	if err := s.storage.Set(ctx, key, ""); err != nil {
		return err
	}

	return nil
}

// Delete - deletes a namespace from the storage.
func (s *NamespaceStorage) Delete(ctx context.Context, namespace string) error {
	key := storage.MakeKey(models.SystemNamespaceNameSpace, namespace)
	if _, err := s.storage.Get(ctx, key); err != nil {
		return ErrNamespaceNotFound
	}

	return s.storage.Del(ctx, key)
}

// Append - adds a new namespace to the list of all namespaces in the system.
func (s *NamespaceStorage) Append(ctx context.Context, namespace string) ([]string, error) {
	nsList, err := s.List(ctx)
	if err != nil && !errors.Is(err, ErrEmptyNamespaces) {
		return nil, err
	}
	nsList = append(nsList, namespace)

	nsBytes, err := gob.Encode(nsList)
	if err != nil {
		return nsList, err
	}

	err = s.storage.Set(ctx, models.SystemNamespacesKey, string(nsBytes))
	if err != nil {
		return nil, err
	}

	return nsList, nil
}

// List - retrieves a list of all namespaces in the system.
func (s *NamespaceStorage) List(ctx context.Context) ([]string, error) {
	namespacesString, err := s.storage.Get(ctx, models.SystemNamespacesKey)
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			return nil, ErrEmptyNamespaces
		}

		return nil, err
	}

	var namespaces []string
	err = gob.Decode([]byte(namespacesString), &namespaces)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	return namespaces, nil
}
