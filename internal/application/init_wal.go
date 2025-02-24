package application

import (
	"github.com/neekrasov/kvdb/internal/config"
	"github.com/neekrasov/kvdb/internal/database/storage/wal"
	"github.com/neekrasov/kvdb/internal/database/storage/wal/compressor"
	"github.com/neekrasov/kvdb/internal/database/storage/wal/filesystem"
	"github.com/neekrasov/kvdb/internal/database/storage/wal/segment"
	sizeparser "github.com/neekrasov/kvdb/pkg/size_parser"
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
		size, err := sizeparser.ParseSize(cfg.MaxSegmentSize)
		if err != nil {
			return nil, err
		}

		segmentManagerOpts = append(segmentManagerOpts, wal.WithMaxSegmentSize(size))
	}
	if cfg.Compression == "gzip" {
		segmentManagerOpts = append(segmentManagerOpts, wal.WithCompressor(new(compressor.GzipCompressor)))
	}

	segmentManager, err := wal.NewFileSegmentManager(
		segmentStorage, segmentManagerOpts...)
	if err != nil {
		return nil, err
	}

	return wal.NewWAL(segmentManager, cfg.FlushingBatchSize, cfg.FlushingBatchTimeout), nil
}
