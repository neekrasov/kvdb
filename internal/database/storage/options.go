package storage

import (
	"time"

	"github.com/neekrasov/kvdb/internal/database/storage/replication"
)

// StorageOpt - options for configuring Storage.
type StorageOpt func(*Storage)

// WithWALOpt - configures Storage with a WAL.
func WithWALOpt(w WAL) StorageOpt {
	return func(s *Storage) {
		s.wal = w
	}
}

// WithReplicaOpt - configures Storage with a replica.
func WithReplicaOpt(r Replica) StorageOpt {
	return func(s *Storage) {
		s.replica = r
	}
}

// WithReplicaStreamOpt - configures Storage with a replica.
func WithReplicaStreamOpt(rs replication.Stream) StorageOpt {
	return func(s *Storage) {
		s.stream = rs
	}
}

// WithPartitionNum - configures Engine with a cleanup period.
func WithCleanupPeriod(cleanupPeriod time.Duration) StorageOpt {
	return func(s *Storage) {
		s.cleanupPeriod = cleanupPeriod
	}
}

// WithPartitionNum - configures Engine with a cleanup period.
func WithCleanupBatchSize(batchSize int) StorageOpt {
	return func(s *Storage) {
		s.cleanupBatchSize = batchSize
	}
}
