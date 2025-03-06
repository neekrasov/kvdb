package wal

import "github.com/neekrasov/kvdb/internal/database/compression"

// FileSegmentManagerOpt - options for configuring FileSegmentManager.
type FileSegmentManagerOpt func(*FileSegmentManager)

// WithCompressor - configures FileSegmentManager with a compression.
func WithCompressor(compression compression.Compressor) FileSegmentManagerOpt {
	return func(fsm *FileSegmentManager) {
		fsm.compression = compression
	}
}

// WithMaxSegmentSize - configures FileSegmentManager with a maximum segment size.
func WithMaxSegmentSize(maxSegmentSize int) FileSegmentManagerOpt {
	return func(fsm *FileSegmentManager) {
		fsm.maxSegmentSize = maxSegmentSize
	}
}
