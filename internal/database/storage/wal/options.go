package wal

// FileSegmentManagerOpt - options for configuring FileSegmentManager.
type FileSegmentManagerOpt func(*FileSegmentManager)

// WithCompressor - configures FileSegmentManager with a compressor.
func WithCompressor(compressor Compressor) func(*FileSegmentManager) {
	return func(fsm *FileSegmentManager) {
		fsm.compressor = compressor
	}
}

// WithMaxSegmentSize - configures FileSegmentManager with a maximum segment size.
func WithMaxSegmentSize(maxSegmentSize int) func(*FileSegmentManager) {
	return func(fsm *FileSegmentManager) {
		fsm.maxSegmentSize = maxSegmentSize
	}
}
