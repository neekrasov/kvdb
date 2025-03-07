package application

import (
	"github.com/neekrasov/kvdb/internal/config"
	"github.com/neekrasov/kvdb/internal/database/compression"
	"github.com/neekrasov/kvdb/internal/database/storage/wal"
	"github.com/neekrasov/kvdb/internal/database/storage/wal/filesystem"
	"github.com/neekrasov/kvdb/internal/database/storage/wal/segment"
	"github.com/neekrasov/kvdb/pkg/sizeutil"
)

func initWAL(cfg *config.WALConfig) (*wal.WAL, error) {
	if cfg == nil {
		return nil, nil
	}

	segmentStorage, err := segment.NewFileSegmentStorage(
		new(filesystem.LocalFileSystem), cfg.DataDir)
	if err != nil {
		return nil, err
	}

	segmentManagerOpts := make([]wal.FileSegmentManagerOpt, 0)
	if cfg.MaxSegmentSize != "" {
		size, err := sizeutil.ParseSize(cfg.MaxSegmentSize)
		if err != nil {
			return nil, err
		}

		segmentManagerOpts = append(segmentManagerOpts, wal.WithMaxSegmentSize(size))
	}
	var compressor compression.Compressor
	if cfg.Compression != "" {
		compressor, err = compression.New(cfg.Compression)
		if err != nil {
			return nil, err
		}
	}

	if cfg.Compression == "gzip" {
		segmentManagerOpts = append(segmentManagerOpts, wal.WithCompressor(compressor))
	}

	segmentManager, err := wal.NewFileSegmentManager(
		segmentStorage, segmentManagerOpts...)
	if err != nil {
		return nil, err
	}

	return wal.NewWAL(segmentManager, cfg.FlushingBatchSize, cfg.FlushingBatchTimeout), nil
}
