package application

import (
	"time"

	"github.com/neekrasov/kvdb/internal/config"
	"github.com/neekrasov/kvdb/internal/database/compression"
	"github.com/neekrasov/kvdb/internal/database/storage/wal"
	"github.com/neekrasov/kvdb/internal/database/storage/wal/filesystem"
	"github.com/neekrasov/kvdb/internal/database/storage/wal/segment"
	"github.com/neekrasov/kvdb/pkg/logger"
	"github.com/neekrasov/kvdb/pkg/sizeutil"
	"go.uber.org/zap"
)

const (
	defaultFlushingBatchTimeout = time.Duration(50)
	defaultDataDir              = "/var/lib/kvdb"
)

func initWAL(cfg *config.WALConfig) (*wal.WAL, error) {
	if cfg == nil {
		logger.Warn("empty wal config")
		return nil, nil
	}

	dataDir := cfg.DataDir
	if cfg.DataDir == "" {
		dataDir = defaultDataDir
	}

	segmentStorage, err := segment.NewFileSegmentStorage(
		new(filesystem.LocalFileSystem), dataDir)
	if err != nil {
		return nil, err
	}

	segmentManagerOpts := make([]wal.FileSegmentManagerOpt, 0)
	maxSegmentSize := defaultMaxSegmentSize
	if cfg.MaxSegmentSize != "" {
		size, err := sizeutil.ParseSize(cfg.MaxSegmentSize)
		if err != nil {
			return nil, err
		}
		maxSegmentSize = size

	}
	segmentManagerOpts = append(
		segmentManagerOpts, wal.WithMaxSegmentSize(maxSegmentSize))

	var compressor compression.Compressor
	if cfg.Compression != "" {
		compressor, err = compression.New(cfg.Compression)
		if err != nil {
			return nil, err
		}
	}
	segmentManagerOpts = append(segmentManagerOpts, wal.WithCompressor(compressor))

	segmentManager, err := wal.NewFileSegmentManager(
		segmentStorage, segmentManagerOpts...)
	if err != nil {
		return nil, err
	}

	var flushingBatchTimeout = defaultFlushingBatchTimeout
	if cfg.FlushingBatchTimeout != 0 {
		flushingBatchTimeout = cfg.FlushingBatchTimeout
	}

	var batchSize = defaultMaxSegmentSize
	if cfg.FlushingBatchSize != 0 {
		batchSize = cfg.FlushingBatchSize
	}

	logger.Debug("init wal",
		zap.Stringer("flushing_batch_timeout", flushingBatchTimeout),
		zap.Int("flushing_batch_size", batchSize),
		zap.String("compression", string(cfg.Compression)),
	)

	return wal.NewWAL(segmentManager, batchSize, flushingBatchTimeout), nil
}
