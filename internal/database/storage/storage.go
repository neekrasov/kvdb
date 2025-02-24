package storage

import (
	"errors"
	"fmt"

	"github.com/neekrasov/kvdb/internal/database/compute"
	"github.com/neekrasov/kvdb/internal/database/storage/replication"
	"github.com/neekrasov/kvdb/internal/database/storage/wal"
	"github.com/neekrasov/kvdb/pkg/logger"
	"go.uber.org/zap"
)

var (
	ErrorMutableOp = errors.New("mutable operation on slave")
	ErrKeyNotFound = errors.New("key not found")
)

type (
	// Engine - key-value storage operations.
	Engine interface {
		Set(key, value string)
		Get(key string) (string, bool)
		Del(key string) error
	}

	// WAL - Write-Ahead Log interface for data persistence.
	WAL interface {
		Set(key, value string) error
		Del(key string) error
		Recover(applyFunc func(entry wal.LogEntry) error) error
	}

	Replica interface {
		IsMaster() bool
	}
)

// Storage - struct that provides a higher-level abstraction
// over the Engine interface for key-value storage operations.
type Storage struct {
	stream  replication.Stream
	replica Replica
	engine  Engine
	wal     WAL
}

// NewStorage - initializes and returns a new Storage instance with the provided storage engine.
func NewStorage(
	engine Engine,
	opts ...StorageOpt,
) (*Storage, error) {
	s := &Storage{engine: engine}
	for _, option := range opts {
		option(s)
	}

	if s.wal != nil {
		if err := s.wal.Recover(s.applyFunc); err != nil {
			return nil, fmt.Errorf("wal recovering failed: %w", err)
		}
	}

	if s.stream != nil {
		go func() {
			for logs := range s.stream {
				for _, log := range logs {
					err := s.applyFunc(log)
					if err != nil {
						logger.Warn("apply operaion failed",
							zap.Int("operation", int(log.Operation)),
							zap.Strings("args", log.Args),
							zap.Error(err),
						)
					}
				}
			}
		}()
	}

	return s, nil
}

// Set - stores a key-value pair in the storage
func (s *Storage) Set(key, value string) error {
	if s.replica != nil && !s.replica.IsMaster() {
		return ErrorMutableOp
	}

	err := s.wal.Set(key, value)
	if err != nil {
		return err
	}

	s.engine.Set(key, value)
	return nil
}

// Get - retrieves the value associated with a key from the storage
func (s *Storage) Get(key string) (string, error) {
	val, exists := s.engine.Get(key)
	if !exists {
		return "", ErrKeyNotFound
	}

	return val, nil
}

// Del - deletes a key-value pair from the storage.
func (s *Storage) Del(key string) error {
	if s.replica != nil && !s.replica.IsMaster() {
		return ErrorMutableOp
	}

	err := s.wal.Del(key)
	if err != nil {
		return err
	}

	return s.engine.Del(key)
}

// MakeKey - constructs a key by combining a namespace and a key name using a colon (:).
func MakeKey(namespace, key string) string {
	return namespace + ":" + key
}

func (s *Storage) applyFunc(entry wal.LogEntry) error {
	switch entry.Operation {
	case compute.SetCommandID:
		s.engine.Set(entry.Args[0], entry.Args[1])
	case compute.DelCommandID:
		err := s.engine.Del(entry.Args[0])
		if err != nil {
			return fmt.Errorf("apply del (%s) failed: %w", entry.Args[0], err)
		}
	case compute.UnknownCommandID:
		return nil
	default:
		return fmt.Errorf("unrecognized command (id: %d, args %v)", entry.Operation, entry.Args)
	}

	logger.Debug("recovered log entry",
		zap.Int("operation", int(entry.Operation)),
		zap.Strings("args", entry.Args),
	)

	return nil
}
