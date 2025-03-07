package wal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/neekrasov/kvdb/internal/database/compression"
	"github.com/neekrasov/kvdb/pkg/logger"
	"go.uber.org/zap"
)

// defaultMaxSegmentSize - is 4KB.
const (
	defaultMaxSegmentSize = 4 << 20
)

type (
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
	compression    compression.Compressor
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
func (fsm *FileSegmentManager) Write(entries []WriteEntry, nolock bool) error {
	fsm.mu.Lock()
	defer fsm.mu.Unlock()

	var buf bytes.Buffer
	for _, entry := range entries {
		log := entry.Log()
		if err := log.Encode(&buf); err != nil {
			err := fmt.Errorf(
				"encode op %d with args %v failed: %w",
				log.Operation, log.Args, err)

			if !nolock {
				fsm.ackEntries(entries, err)
			}

			return err
		}
	}

	defaultOffset := fsm.maxSegmentSize / 10
	if fsm.current != nil && fsm.current.Size()+defaultOffset+buf.Len() > fsm.maxSegmentSize {
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

	if !nolock {
		fsm.ackEntries(entries, nil)
	}

	return nil
}

// rotate - rotates to a new segment.
func (fsm *FileSegmentManager) rotate() error {
	var sID int
	if fsm.current != nil {
		if err := fsm.current.Close(); err != nil {
			return fmt.Errorf("failed to close current segment: %w", err)
		}

		if fsm.compression != nil {
			if err := fsm.compress(fsm.current.ID()); err != nil {
				logger.Error("failed to compress segment", zap.Int("segment_id", fsm.current.ID()), zap.Error(err))
			}
		}
		sID = fsm.current.ID() + 1
	} else {
		sID = 1
	}

	writer, err := fsm.storage.Create(sID, false)
	if err != nil {
		return fmt.Errorf("failed to create new segment: %w", err)
	}

	fsm.current = writer
	fsm.segments = append(fsm.segments, sID)
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

	compressed, err := fsm.compression.Compress(data)
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
func (fsm *FileSegmentManager) ForEach(action func(context.Context, []byte) error) error {
	if action == nil {
		return nil
	}

	iterator := NewSegmentIterator(fsm.storage, fsm.compression)
	for _, n := range fsm.segments {
		data, err := iterator.Next(n)
		if err != nil {
			if err == io.EOF {
				break
			}

			return fmt.Errorf("iteration failed: %w", err)
		}

		if err := action(context.TODO(), data); err != nil {
			return fmt.Errorf("action failed (s.num %d): %w", n, err)
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
