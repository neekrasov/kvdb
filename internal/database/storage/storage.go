package storage

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/neekrasov/kvdb/internal/database/compute"
	"github.com/neekrasov/kvdb/internal/database/storage/replication"
	"github.com/neekrasov/kvdb/internal/database/storage/wal"
	"github.com/neekrasov/kvdb/pkg/ctxutil"
	"github.com/neekrasov/kvdb/pkg/logger"
	pkgsync "github.com/neekrasov/kvdb/pkg/sync"
	"go.uber.org/zap"
)

var (
	ErrorMutableOp = errors.New("mutable operation on slave")
	ErrKeyNotFound = errors.New("key not found")
)

type (
	// Stats - structure for storing 'storage' statistics.
	Stats struct {
		StartTime     time.Time    `json:"start_time"`     // Server startup time.
		TotalCommands atomic.Int64 `json:"total_commands"` // Total number of commands executed.
		GetCommands   atomic.Int64 `json:"get_commands"`   // Number of GET commands.
		SetCommands   atomic.Int64 `json:"set_commands"`   // Number of SET commands.
		DelCommands   atomic.Int64 `json:"del_commands"`   // Number of DEL commands.
		TotalKeys     atomic.Int64 `json:"total_keys"`     // Total number of keys in the storage (approximate).
		ExpiredKeys   atomic.Int64 `json:"expired_keys"`   // Number of expired keys (deleted).
	}

	// Engine - key-value storage operations.
	Engine interface {
		Set(ctx context.Context, key, value string, ttl int64)
		Get(ctx context.Context, key string) (string, bool)
		Del(ctx context.Context, key string) error
		Watch(ctx context.Context, key string) pkgsync.FutureString
		ForEachExpired(action func(key string))
	}

	// WAL - Write-Ahead Log interface for data persistence.
	WAL interface {
		Set(ctx context.Context, key, value string) error
		Del(ctx context.Context, key string) error
		Recover(applyFunc func(ctx context.Context, entry []wal.LogEntry) error) (int64, error)
		Flush(batch []wal.WriteEntry) error
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

	stats *Stats

	cleanupPeriod    time.Duration
	cleanupBatchSize int
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

	if s.cleanupPeriod != 0 &&
		(s.replica == nil || s.replica.IsMaster()) {
		go s.startCleanupExpiresKeys(ctx)
	}

	return s, nil
}

// Set - stores a key-value pair in the storage
func (s *Storage) Set(ctx context.Context, key, value string) error {
	if s.replica != nil && !s.replica.IsMaster() {
		return ErrorMutableOp
	}

	var ttl int64
	if ttlStr := ctxutil.ExtractTTL(ctx); ttlStr != "" {
		duration, err := time.ParseDuration(ttlStr)
		if err != nil {
			return fmt.Errorf("invalid format fo ttl: %w", err)
		}

		ttl = time.Now().Unix() + (duration.Nanoseconds() / 1e9)
	}

	ctx = ctxutil.InjectTxID(ctx, s.gen.Generate())
	err := s.wal.Set(ctx, key, value)
	if err != nil {
		return err
	}

	if s.stats != nil {
		s.stats.SetCommands.Add(1)
		s.stats.TotalCommands.Add(1)

		if _, exists := s.engine.Get(ctx, key); !exists {
			s.stats.TotalKeys.Add(1)
		}
	}

	s.engine.Set(ctx, key, value, ttl)

	return nil
}

// Get - retrieves the value associated with a key from the storage
func (s *Storage) Get(ctx context.Context, key string) (string, error) {
	txID := s.gen.Generate()
	ctx = ctxutil.InjectTxID(ctx, txID)

	val, exists := s.engine.Get(ctx, key)
	if !exists {
		return "", ErrKeyNotFound
	}

	if s.stats != nil {
		s.stats.GetCommands.Add(1)
		s.stats.TotalCommands.Add(1)
	}

	return val, nil
}

// Del - deletes a key-value pair from the storage.
func (s *Storage) Del(ctx context.Context, key string) error {
	if s.replica != nil && !s.replica.IsMaster() {
		return ErrorMutableOp
	}

	txID := s.gen.Generate()
	ctx = ctxutil.InjectTxID(ctx, txID)

	err := s.wal.Del(ctx, key)
	if err != nil {
		return err
	}

	err = s.engine.Del(ctx, key)
	if err != nil {
		return err
	}

	if s.stats != nil {
		s.stats.DelCommands.Add(1)
		s.stats.TotalCommands.Add(1)
		s.stats.TotalKeys.Add(-1)
	}

	return nil
}

// Watch - watches the key and returns the value if it has changed.
func (s *Storage) Watch(ctx context.Context, key string) pkgsync.FutureString {
	txID := s.gen.Generate()
	ctx = ctxutil.InjectTxID(ctx, txID)

	return s.engine.Watch(ctx, key)
}

// MakeKey - constructs a key by combining a namespace and a key name using a colon (:).
func MakeKey(namespace, key string) string {
	return namespace + ":" + key
}

func (s *Storage) applyFunc(ctx context.Context, entries []wal.LogEntry) error {
	var lastLSN int64
	for _, entry := range entries {
		lastLSN = max(lastLSN, entry.LSN)
		ctx := ctxutil.InjectTxID(ctx, entry.LSN)

		switch entry.Operation {
		case compute.SetCommandID:
			s.engine.Set(ctx, entry.Args[0], entry.Args[1], 0)

			if s.stats != nil {
				s.stats.SetCommands.Add(1)
			}
		case compute.DelCommandID:
			if err := s.engine.Del(ctx, entry.Args[0]); err != nil {
				return fmt.Errorf("apply del (%s) failed: %w", entry.Args[0], err)
			}

			if s.stats != nil {
				s.stats.DelCommands.Add(1)
			}
		case compute.UnknownCommandID:
			return nil
		default:
			return fmt.Errorf("unrecognized command (id: %d, args %v)", entry.Operation, entry.Args)
		}

		if s.stats != nil {
			s.stats.TotalCommands.Add(1)
		}

		logger.Debug("recovered log entry",
			zap.Int64("lsn", entry.LSN),
			zap.Int("operation", int(entry.Operation)),
			zap.Strings("args", entry.Args),
		)
	}

	return nil
}

func (s *Storage) startCleanupExpiresKeys(ctx context.Context) {
	ticker := time.NewTicker(s.cleanupPeriod)
	defer ticker.Stop()

	entries := make([]wal.WriteEntry, 0, s.cleanupBatchSize)
	for {
		select {
		case <-ticker.C:
			logger.Debug("start removing expires keys")

			s.engine.ForEachExpired(func(key string) {
				entries = append(entries, wal.NewWriteEntry(
					s.gen.Generate(), compute.DelCommandID, []string{key},
				))
				if len(entries) == s.cleanupBatchSize {
					s.cleanupKeys(ctx, entries)
					entries = entries[:0]
				}
			})

			if len(entries) > 0 {
				s.cleanupKeys(ctx, entries)
				entries = entries[:0]
			}
		case <-ctx.Done():
			logger.Debug("cleanup expired key stopped", zap.Stringer("time", time.Now().UTC()))
			return
		}
	}
}

func (s *Storage) cleanupKeys(ctx context.Context, entries []wal.WriteEntry) {
	s.wal.Flush(entries)
	for _, entry := range entries {
		key := entry.Log().Args[0]
		s.engine.Del(ctx, key)
		if s.stats != nil {
			s.stats.ExpiredKeys.Add(1)
		}
		logger.Debug("removed expired key (background)", zap.String("key", key))
	}
}

// Stats - returns the collected database statistics.
func (s *Storage) Stats() (*Stats, error) {
	if s.stats == nil {
		return nil, errors.New("statistics disabled")
	}

	return s.stats, nil
}
