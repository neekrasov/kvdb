package storage

import (
	"sync"

	"github.com/neekrasov/kvdb/internal/database"
)

// InMemoryEngine is a simple in-memory implementation of Engine
// It uses a map with a mutex for thread-safe operations.
type InMemoryEngine struct {
	data map[string]string
	mu   sync.RWMutex
}

// NewInMemoryEngine - creates a new instance of InMemoryEngine.
func NewInMemoryEngine() *InMemoryEngine {
	return &InMemoryEngine{
		data: make(map[string]string),
	}
}

// Set - set stores a key-value pair in memory.
func (e *InMemoryEngine) Set(key, value string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.data[key] = value
	return nil
}

// Get - retrieves the value associated with a key.
func (e *InMemoryEngine) Get(key string) (string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	value, exists := e.data[key]
	if !exists {
		return "", database.ErrKeyNotFound
	}

	return value, nil
}

// Del - removes a key-value pair from memory.
func (e *InMemoryEngine) Del(key string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.data[key]; !exists {
		return database.ErrKeyNotFound
	}
	delete(e.data, key)

	return nil
}
