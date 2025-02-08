package storage

import (
	"github.com/neekrasov/kvdb/internal/database/storage/models"
)

type Engine interface {
	Set(key, value string)
	Get(key string) (string, bool)
	Del(key string) error
}

// Storage - A struct that provides a higher-level abstraction
// over the Engine interface for key-value storage operations.
type Storage struct {
	engine Engine
}

// NewStorage - Initializes and returns a new Storage instance with the provided storage engine.
func NewStorage(
	engine Engine,
) *Storage {
	return &Storage{
		engine: engine,
	}
}

// Set - Stores a key-value pair in the storage
func (s *Storage) Set(key, value string) {
	s.engine.Set(key, value)
}

// Get - Retrieves the value associated with a key from the storage
func (s *Storage) Get(key string) (string, error) {
	val, exists := s.engine.Get(key)
	if !exists {
		return "", models.ErrKeyNotFound
	}

	return val, nil
}

// Del - Deletes a key-value pair from the storage.
func (s *Storage) Del(key string) error {
	return s.engine.Del(key)
}

// MakeKey - Constructs a key by combining a namespace and a key name using a colon (:).
func MakeKey(namespace, key string) string {
	return namespace + ":" + key
}
