package wal_test

import (
	"errors"
	"io"
	"sync"
	"testing"

	"github.com/neekrasov/kvdb/internal/database/compute"
	"github.com/neekrasov/kvdb/internal/database/storage/wal"
	"github.com/neekrasov/kvdb/internal/database/storage/wal/segment"
	mocks "github.com/neekrasov/kvdb/internal/mocks/wal"
	"github.com/neekrasov/kvdb/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewFileSegmentManager(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	tests := []struct {
		name         string
		prepareMocks func(mockStorage *mocks.SegmentStorage)
		expectError  bool
	}{
		{
			name: "Success - Create new manager with existing segments",
			prepareMocks: func(mockStorage *mocks.SegmentStorage) {
				mockStorage.EXPECT().List().Return([]int{1, 2}, nil)
			},
			expectError: false,
		},
		{
			name: "Success - Create new manager with no segments",
			prepareMocks: func(mockStorage *mocks.SegmentStorage) {
				mockStorage.EXPECT().List().Return([]int{}, nil)
				mockStorage.EXPECT().Create(1, false).Return(&segment.Segment{}, nil)
			},
			expectError: false,
		},
		{
			name: "Error - Failed to list segments",
			prepareMocks: func(mockStorage *mocks.SegmentStorage) {
				mockStorage.EXPECT().List().Return(nil, errors.New("list error"))
			},
			expectError: true,
		},
		{
			name: "Error - Failed to create initial segment",
			prepareMocks: func(mockStorage *mocks.SegmentStorage) {
				mockStorage.EXPECT().List().Return([]int{}, nil)
				mockStorage.EXPECT().Create(1, false).Return(nil, errors.New("create error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewSegmentStorage(t)
			tt.prepareMocks(mockStorage)

			manager, err := wal.NewFileSegmentManager(mockStorage)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, manager)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, manager)
			}
		})
	}
}

func TestFileSegmentManager_Write(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	tests := []struct {
		name         string
		entries      []wal.WriteEntry
		prepareMocks func(mockStorage *mocks.SegmentStorage, mockSegment *mocks.Segment)
		expectError  bool
	}{
		{
			name: "Success - Write entries to current segment",
			entries: []wal.WriteEntry{
				wal.NewWriteEntry(compute.SetCommandID, []string{}),
			},
			prepareMocks: func(mockStorage *mocks.SegmentStorage, mockSegment *mocks.Segment) {
				mockStorage.EXPECT().List().Return([]int{1}, nil)
				mockStorage.EXPECT().Create(1, false).Return(mockSegment, nil)
				mockSegment.EXPECT().ID().Return(1)
				mockSegment.EXPECT().Write(mock.Anything).Return(0, nil)
			},
			expectError: false,
		},
		{
			name: "Error - Failed to write to segment",
			entries: []wal.WriteEntry{
				wal.NewWriteEntry(compute.SetCommandID, []string{}),
			},
			prepareMocks: func(mockStorage *mocks.SegmentStorage, mockSegment *mocks.Segment) {
				mockStorage.EXPECT().List().Return([]int{1}, nil)
				mockStorage.EXPECT().Create(1, false).Return(mockSegment, nil)
				mockSegment.EXPECT().ID().Return(1)
				mockSegment.EXPECT().Write(mock.Anything).Return(0, errors.New("write error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewSegmentStorage(t)
			mockSegment := mocks.NewSegment(t)
			tt.prepareMocks(mockStorage, mockSegment)

			manager, err := wal.NewFileSegmentManager(mockStorage)
			require.NoError(t, err)

			w := new(sync.WaitGroup)
			w.Add(len(tt.entries))
			for _, entry := range tt.entries {
				go func() {
					w.Done()
					err = entry.Get()
					require.NoError(t, err)
				}()
			}

			err = manager.Write(tt.entries, false)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			w.Wait()

			manager.SetCurrent(nil)
			err = manager.Close()
			require.NoError(t, err)
		})
	}
}

func TestFileSegmentManager_Write_WithCompression(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	tests := []struct {
		name         string
		entries      []wal.WriteEntry
		prepareMocks func(mockStorage *mocks.SegmentStorage, mockSegment *mocks.Segment, mockCompressor *mocks.Compressor)
		expectError  bool
	}{
		{
			name: "Success - Write entries with compression",
			entries: []wal.WriteEntry{
				wal.NewWriteEntry(compute.SetCommandID, []string{"key", "value"}),
			},
			prepareMocks: func(mockStorage *mocks.SegmentStorage, mockSegment *mocks.Segment, mockCompressor *mocks.Compressor) {
				mockStorage.EXPECT().List().Return([]int{1}, nil)
				mockStorage.EXPECT().Open(1).Return(mockSegment, nil)
				mockSegment.EXPECT().ID().Return(1)
				mockSegment.EXPECT().Size().Return(4 << 10)
				mockSegment.EXPECT().Close().Return(nil)

				mockSegment.EXPECT().Read(mock.Anything).Return(0, io.EOF)
				mockStorage.EXPECT().Remove(1).Return(nil)
				mockStorage.EXPECT().Create(1, true).Return(mockSegment, nil)
				mockStorage.EXPECT().Create(2, false).Return(mockSegment, nil)
				mockCompressor.EXPECT().Compress(mock.Anything).Return([]byte("compressed_data"), nil)
				mockSegment.EXPECT().Write([]byte("compressed_data")).Return(len([]byte("compressed_data")), nil)
				mockSegment.EXPECT().Write(mock.Anything).Return(0, nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewSegmentStorage(t)
			mockSegment := mocks.NewSegment(t)
			mockCompressor := mocks.NewCompressor(t)
			tt.prepareMocks(mockStorage, mockSegment, mockCompressor)

			manager, err := wal.NewFileSegmentManager(mockStorage,
				wal.WithCompressor(mockCompressor), wal.WithMaxSegmentSize(1))
			require.NoError(t, err)

			manager.SetCurrent(mockSegment)

			w := new(sync.WaitGroup)
			w.Add(len(tt.entries))
			for _, entry := range tt.entries {
				go func(e wal.WriteEntry) {
					defer w.Done()
					err = e.Get()
					require.NoError(t, err)
				}(entry)
			}

			err = manager.Write(tt.entries, false)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			w.Wait()

			err = manager.Close()
			require.NoError(t, err)
		})
	}
}

func TestFileSegmentManager_ForEach(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	tests := []struct {
		name         string
		prepareMocks func(mockStorage *mocks.SegmentStorage, mockSegment *mocks.Segment, mockCompressor *mocks.Compressor)
		expectError  bool
		action       func([]byte) error
	}{
		{
			name: "Success - Iterate over segments",
			prepareMocks: func(mockStorage *mocks.SegmentStorage, mockSegment *mocks.Segment, mockCompressor *mocks.Compressor) {
				mockStorage.EXPECT().List().Return([]int{1, 2}, nil)

				mockStorage.EXPECT().Open(1).Return(mockSegment, nil)
				mockSegment.EXPECT().Close().Return(nil)
				mockSegment.EXPECT().Read(mock.Anything).Return(0, io.EOF)
				mockSegment.EXPECT().Compressed().Return(false).Once()

				mockStorage.EXPECT().Open(2).Return(mockSegment, nil)
				mockSegment.EXPECT().Close().Return(nil)
				mockSegment.EXPECT().Read(mock.Anything).Return(0, io.EOF)
				mockSegment.EXPECT().Compressed().Return(true).Once()
				mockCompressor.EXPECT().Decompress([]uint8{}).Return(nil, nil)
			},
			action: func(b []byte) error {
				return nil
			},
			expectError: false,
		},
		{
			name: "Error - Failed to decompress segment",
			prepareMocks: func(mockStorage *mocks.SegmentStorage, mockSegment *mocks.Segment, mockCompressor *mocks.Compressor) {
				mockStorage.EXPECT().List().Return([]int{1, 2}, nil)

				mockStorage.EXPECT().Open(1).Return(mockSegment, nil)
				mockSegment.EXPECT().Close().Return(nil)
				mockSegment.EXPECT().Read(mock.Anything).Return(0, io.EOF)
				mockSegment.EXPECT().Compressed().Return(false).Once()

				mockStorage.EXPECT().Open(2).Return(mockSegment, nil)
				mockSegment.EXPECT().Close().Return(nil)
				mockSegment.EXPECT().Read(mock.Anything).Return(0, io.EOF)
				mockSegment.EXPECT().Compressed().Return(true).Once()
				mockCompressor.EXPECT().Decompress([]uint8{}).Return(nil, errors.New("some error"))
			},
			action: func(b []byte) error {
				return nil
			},
			expectError: true,
		},
		{
			name: "Error - Openning error",
			prepareMocks: func(mockStorage *mocks.SegmentStorage, mockSegment *mocks.Segment, mockCompressor *mocks.Compressor) {
				mockStorage.EXPECT().List().Return([]int{1}, nil)
				mockStorage.EXPECT().Open(1).Return(mockSegment, errors.New("opening error"))
			},
			action: func(b []byte) error {
				return nil
			},
			expectError: true,
		},
		{
			name: "Error - Error while reading segment",
			prepareMocks: func(mockStorage *mocks.SegmentStorage, mockSegment *mocks.Segment, mockCompressor *mocks.Compressor) {
				mockStorage.EXPECT().List().Return([]int{1}, nil)
				mockStorage.EXPECT().Open(1).Return(mockSegment, nil)
				mockSegment.EXPECT().Close().Return(nil)
				mockSegment.EXPECT().Read(mock.Anything).Return(0, errors.New("test"))
			},
			action: func(b []byte) error {
				return nil
			},
			expectError: true,
		},
		{
			name: "Success - Nil Action",
			prepareMocks: func(mockStorage *mocks.SegmentStorage, mockSegment *mocks.Segment, mockCompressor *mocks.Compressor) {
				mockStorage.EXPECT().List().Return([]int{1}, nil)
			},
			action:      nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewSegmentStorage(t)
			mockSegment := mocks.NewSegment(t)
			mockCompressor := mocks.NewCompressor(t)
			tt.prepareMocks(mockStorage, mockSegment, mockCompressor)

			manager, err := wal.NewFileSegmentManager(mockStorage, wal.WithCompressor(mockCompressor))
			require.NoError(t, err)

			err = manager.ForEach(tt.action)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
