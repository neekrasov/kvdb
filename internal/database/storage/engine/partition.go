package engine

import (
	"sync"

	"github.com/neekrasov/kvdb/internal/database"
)

// partitionMap - represents one data partition.
type partitionMap struct {
	data map[string]string
	mu   sync.RWMutex
}

// newPartMap - returns a new partition instance
func newPartMap() *partitionMap {
	return &partitionMap{
		data: make(map[string]string),
	}
}

// Set - set stores a key-value pair in memory.
func (p *partitionMap) Set(key, value string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.data[key] = value
}

// Get - retrieves the value associated with a key.
func (p *partitionMap) Get(key string) (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	val, exists := p.data[key]
	return val, exists
}

// Del - removes a key-value pair from memory.
func (p *partitionMap) Del(key string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.data[key]; !exists {
		return database.ErrKeyNotFound
	}
	delete(p.data, key)

	return nil
}
