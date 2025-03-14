package wal

import (
	"errors"
	"fmt"
	"io"

	"github.com/neekrasov/kvdb/internal/database/compression"
)

// ErrSegmentNotFound - error returned when the requested segment is not found in storage.
var ErrSegmentNotFound = errors.New("segment not found")

// SegmentIterator - struct for iterating over stored segments, with optional decompression support.
type SegmentIterator struct {
	storage     SegmentStorage
	compression compression.Compressor
}

// NewSegmentIterator - constructor function that creates a new SegmentIterator.
func NewSegmentIterator(storage SegmentStorage, compression compression.Compressor) *SegmentIterator {
	return &SegmentIterator{
		storage:     storage,
		compression: compression,
	}
}

// Next - retrieves the next segment's data, decompressing it if necessary.
// Returns io.EOF if the segment is not found.
func (si *SegmentIterator) Next(num int) ([]byte, error) {
	nums, err := si.storage.List()
	if err != nil {
		return nil, err
	}

	seg, err := si.storage.Open(num)
	if err != nil {
		return nil, fmt.Errorf("failed to open segment %d: %w", num, err)
	}
	defer seg.Close()

	data, err := io.ReadAll(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to read segment %d: %w", num, err)
	}

	if seg.Compressed() {
		if si.compression == nil {
			return nil, errors.New("compression not initialized, cannot decompress segment")
		}

		data, err = si.compression.Decompress(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress segment %d: %w", num, err)
		}
	}

	if num == nums[len(nums)-1] {
		return data, io.EOF
	}

	return data, nil
}
