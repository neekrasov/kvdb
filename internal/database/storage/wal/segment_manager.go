package wal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"
	"sync"

	"github.com/neekrasov/kvdb/pkg/logger"
	"go.uber.org/zap"
)

// defaultMaxSegmentSize - is 4KB.
const defaultMaxSegmentSize = 4 << 10

type (
	// Compressor - interface for data compression and decompression.
	Compressor interface {
		// Compress - compresses input data ([]byte) using some compress algorithm.
		Compress(data []byte) ([]byte, error)
		// Decompress - decompresses compressed data ([]byte) compressed using some compress algorithm.
		Decompress(data []byte) ([]byte, error)
	}

	// SegmentStorage - interface for managing storage segments.
	SegmentStorage interface {
		// Create - creates a new segment file.
		Create(id int, compressed bool) (Segment, error)
		// Open - opens an existing segment file.
		Open(id int) (Segment, error)
		// Remove - removes a segment file.
		Remove(id int) error
		// List - lists all segment IDs in the storage.
		List() ([]int, error)
	}

	// Segment - interface for a single storage segment.
	Segment interface {
		// Close - closes the segment file.
		Close() error
		// Compressed - returns whether the segment is compressed.
		Compressed() bool
		// ID - returns the ID of the segment.
		ID() int
		// Read - reads data from the segment file.
		Read(p []byte) (n int, err error)
		// Size - returns the size of the segment.
		Size() int
		// Write - writes data to the segment file.
		Write(data []byte) (int, error)
	}
)

// FileSegmentManager - manages file-based storage segments.
type FileSegmentManager struct {
	storage        SegmentStorage
	compressor     Compressor
	maxSegmentSize int

	mu       sync.Mutex
	current  Segment
	segments []int
}

// NewFileSegmentManager - initializes and returns a new FileSegmentManager.
func NewFileSegmentManager(storage SegmentStorage, opts ...FileSegmentManagerOpt) (*FileSegmentManager, error) {
	segments, err := storage.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list segments: %w", err)
	}

	fsm := &FileSegmentManager{
		storage:  storage,
		segments: segments,
	}

	for _, option := range opts {
		option(fsm)
	}

	if fsm.maxSegmentSize == 0 {
		fsm.maxSegmentSize = defaultMaxSegmentSize
	}

	if len(segments) == 0 {
		if err := fsm.rotate(); err != nil {
			return nil, fmt.Errorf("failed to create initial segment: %w", err)
		}
	}

	return fsm, nil
}

// Write - writes entries to the current segment.
func (fsm *FileSegmentManager) Write(entries []WriteEntry) error {
	fsm.mu.Lock()
	defer fsm.mu.Unlock()

	var buf bytes.Buffer
	for _, entry := range entries {
		log := entry.Log()
		if err := log.Encode(&buf); err != nil {
			err := fmt.Errorf(
				"encode op %d with args %v failed: %w",
				log.Operation, log.Args, err)
			fsm.ackEntries(entries, err)
			return err
		}
	}

	if fsm.current != nil && fsm.current.Size()+buf.Len() > fsm.maxSegmentSize {
		logger.Debug("rotate segment",
			zap.Int("size", fsm.current.Size()),
			zap.Int("id", fsm.current.ID()))

		if err := fsm.rotate(); err != nil {
			return fmt.Errorf("failed to rotate segment: %w", err)
		}
	}

	if fsm.current == nil {
		segment, err := fsm.storage.Create(fsm.segments[len(fsm.segments)-1], false)
		if err != nil {
			return err
		}

		fsm.current = segment
	}

	logger.Debug("write data to segment",
		zap.Int("size_bytes", buf.Len()),
		zap.Int("id", fsm.current.ID()))

	if _, err := fsm.current.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("failed to write to segment: %w", err)
	}

	fsm.ackEntries(entries, nil)
	return nil
}

// rotate - rotates to a new segment.
func (fsm *FileSegmentManager) rotate() error {
	var sID int
	if length := len(fsm.segments); length > 0 {
		sID = fsm.segments[length-1]
	} else {
		sID = 1
	}

	if fsm.current != nil {
		if err := fsm.current.Close(); err != nil {
			return fmt.Errorf("failed to close current segment: %w", err)
		}

		if fsm.compressor != nil {
			if err := fsm.compress(sID); err != nil {
				logger.Error("failed to compress segment", zap.Int("segment_id", sID), zap.Error(err))
			}
		}
	}

	writer, err := fsm.storage.Create(sID, false)
	if err != nil {
		return fmt.Errorf("failed to create new segment: %w", err)
	}

	fsm.current = writer
	fsm.segments = append(fsm.segments, sID+1)
	return nil
}

// compress - compresses a segment.
func (fsm *FileSegmentManager) compress(id int) error {
	logger.Debug("compress segment",
		zap.Int("size", fsm.current.Size()),
		zap.Int("id", fsm.current.ID()))

	reader, err := fsm.storage.Open(id)
	if err != nil {
		return fmt.Errorf("failed to open segment %d: %w", id, err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read segment %d: %w", id, err)
	}

	logger.Debug("remove old no compressed segment",
		zap.Int("size", fsm.current.Size()),
		zap.Int("id", fsm.current.ID()))

	if err := fsm.storage.Remove(id); err != nil {
		return fmt.Errorf("failed to delete uncompressed segment %d: %w", id, err)
	}

	segment, err := fsm.storage.Create(id, true)
	if err != nil {
		return err
	}

	compressed, err := fsm.compressor.Compress(data)
	if err != nil {
		return fmt.Errorf("failed to compress segment %d: %w", id, err)
	}

	if _, err := segment.Write(compressed); err != nil {
		return err
	}

	logger.Debug("writed new compressed segment",
		zap.Int("size", fsm.current.Size()),
		zap.Int("id", fsm.current.ID()))

	return nil
}

// ackEntries - acknowledges written entries.
func (fsm *FileSegmentManager) ackEntries(entries []WriteEntry, err error) {
	for _, entry := range entries {
		entry.Set(err)
	}
}

// ForEach - iterates through all segments.
func (fsm *FileSegmentManager) ForEach(action func([]byte) error) error {
	if action == nil {
		return nil
	}

	fsm.mu.Lock()
	defer fsm.mu.Unlock()

	logger.Debug("for each segments", zap.Ints("segments", fsm.segments))

	if !slices.Contains(fsm.segments, 1) {
		return errors.New("cannot find first segment")
	}

	for _, id := range fsm.segments {
		segment, err := fsm.storage.Open(id)
		if err != nil {
			return fmt.Errorf("failed to open segment %d: %w", id, err)
		}
		defer segment.Close()

		data, err := io.ReadAll(segment)
		if err != nil {
			return fmt.Errorf("read segment for decompress failed: %w", err)
		}

		if segment.Compressed() {
			if fsm.compressor == nil {
				return errors.New("error decompress compressed segment, compressor not initialized")
			}

			logger.Debug("decompress segment",
				zap.Int("id", id),
				zap.Int("size", segment.Size()))
			data, err = fsm.compressor.Decompress(data)
			if err != nil {
				return fmt.Errorf("decompress segment failed: %w", err)
			}
		}

		if err := action(data); err != nil {
			return fmt.Errorf("execute for each action failed: %w", err)
		}
	}

	return nil
}

// Close - closes the current segment.
func (fsm *FileSegmentManager) Close() error {
	fsm.mu.Lock()
	defer fsm.mu.Unlock()

	if fsm.current != nil {
		return fsm.current.Close()
	}

	return nil
}

// SetCurrent - sets the current segment (for testing).
func (fsm *FileSegmentManager) SetCurrent(s Segment) {
	fsm.current = s
}
