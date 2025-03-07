package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/neekrasov/kvdb/internal/database/compute"
	"github.com/neekrasov/kvdb/internal/database/storage/replication"
	"github.com/neekrasov/kvdb/internal/database/storage/tx"
	"github.com/neekrasov/kvdb/internal/database/storage/wal"
	"github.com/neekrasov/kvdb/pkg/logger"
	pkgsync "github.com/neekrasov/kvdb/pkg/sync"
	"go.uber.org/zap"
)

var (
	ErrorMutableOp = errors.New("mutable operation on slave")
	ErrKeyNotFound = errors.New("key not found")
)

type (
	// Engine - key-value storage operations.
	Engine interface {
		Set(ctx context.Context, key, value string)
		Get(ctx context.Context, key string) (string, bool)
		Del(ctx context.Context, key string) error
	}

	// WAL - Write-Ahead Log interface for data persistence.
	WAL interface {
		Set(ctx context.Context, key, value string) error
		Del(ctx context.Context, key string) error
		Recover(applyFunc func(ctx context.Context, entry []wal.LogEntry) error) (int64, error)
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
	gen     *pkgsync.IDGenerator
}

// NewStorage - initializes and returns a new Storage instance with the provided storage engine.
func NewStorage(
	ctx context.Context,
	engine Engine,
	opts ...StorageOpt,
) (*Storage, error) {
	s := &Storage{engine: engine}
	for _, option := range opts {
		option(s)
	}

	var lastLSN int64
	if s.wal != nil {
		var err error
		lastLSN, err = s.wal.Recover(s.applyFunc)
		if err != nil {
			return nil, fmt.Errorf("wal recovering failed: %w", err)
		}
	}

	if s.stream != nil {
		go func() {
			for logs := range s.stream {
				err := s.applyFunc(ctx, logs)
				if err != nil {
					logger.Warn("apply logs batch failed", zap.Error(err))
				}
			}
		}()
	}

	s.gen = pkgsync.NewIDGenerator(lastLSN)
	return s, nil
}

// Set - stores a key-value pair in the storage
func (s *Storage) Set(ctx context.Context, key, value string) error {
	if s.replica != nil && !s.replica.IsMaster() {
		return ErrorMutableOp
	}

	txID := s.gen.Generate()
	ctx = tx.InjectTxID(ctx, txID)

	err := s.wal.Set(ctx, key, value)
	if err != nil {
		return err
	}

	s.engine.Set(ctx, key, value)
	return nil
}

// Get - retrieves the value associated with a key from the storage
func (s *Storage) Get(ctx context.Context, key string) (string, error) {
	txID := s.gen.Generate()
	ctx = tx.InjectTxID(ctx, txID)

	val, exists := s.engine.Get(ctx, key)
	if !exists {
		return "", ErrKeyNotFound
	}

	return val, nil
}

// Del - deletes a key-value pair from the storage.
func (s *Storage) Del(ctx context.Context, key string) error {
	if s.replica != nil && !s.replica.IsMaster() {
		return ErrorMutableOp
	}

	txID := s.gen.Generate()
	ctx = tx.InjectTxID(ctx, txID)

	err := s.wal.Del(ctx, key)
	if err != nil {
		return err
	}

	return s.engine.Del(ctx, key)
}

// MakeKey - constructs a key by combining a namespace and a key name using a colon (:).
func MakeKey(namespace, key string) string {
	return namespace + ":" + key
}

func (s *Storage) applyFunc(ctx context.Context, entries []wal.LogEntry) error {
	for _, entry := range entries {
		switch entry.Operation {
		case compute.SetCommandID:
			s.engine.Set(ctx, entry.Args[0], entry.Args[1])
		case compute.DelCommandID:
			err := s.engine.Del(ctx, entry.Args[0])
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
	}

	return nil
}
