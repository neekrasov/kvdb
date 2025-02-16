package engine

import (
	"hash/fnv"

	"github.com/neekrasov/kvdb/pkg/logger"
	"go.uber.org/zap"
)

// Engine - abstract data storage engine.
type Engine struct {
	partitions []*partitionMap
}

// New - creates a new instance of Engine.
func New(options ...Option) *Engine {
	e := new(Engine)

	for _, option := range options {
		option(e)
	}

	if len(e.partitions) == 0 {
		e.partitions = make([]*partitionMap, 1)
		e.partitions[0] = newPartMap()
	}

	return e
}

// Set - set stores a key-value pair in memory.
func (e *Engine) Set(key, value string) {
	e.part(key).Set(key, value)
}

// Get - retrieves the value associated with a key.
func (e *Engine) Get(key string) (string, bool) {
	return e.part(key).Get(key)
}

// Del - removes a key-value pair from memory.
func (e *Engine) Del(key string) error {
	return e.part(key).Del(key)
}

// part - returns the partition for a given key based on hashing.
func (e *Engine) part(key string) *partitionMap {
	hash := fnv.New32a()
	if _, err := hash.Write([]byte(key)); err != nil {
		logger.Error(
			"hash key failed",
			zap.String("key", key),
			zap.Error(err),
		)
		return nil
	}
	return e.partitions[int(hash.Sum32())%len(e.partitions)]
}
