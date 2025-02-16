package segment_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/neekrasov/kvdb/internal/database/storage/wal/segment"
	mocks "github.com/neekrasov/kvdb/internal/mocks/segment"
	"github.com/neekrasov/kvdb/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileSegmentStorage(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	tests := []struct {
		name         string
		dataDir      string
		prepareMocks func(mockFS *mocks.FileSystem, dataDir string)
		expectError  bool
	}{
		{
			name:    "Success - Directory created",
			dataDir: "/data",
			prepareMocks: func(mockFS *mocks.FileSystem, dataDir string) {
				mockFS.EXPECT().Stat(dataDir).Return(nil, os.ErrNotExist)
				mockFS.EXPECT().MkdirAll(dataDir, os.FileMode(0755)).Return(nil)
			},
			expectError: false,
		},
		{
			name:    "Success - Directory already exists",
			dataDir: "/data",
			prepareMocks: func(mockFS *mocks.FileSystem, dataDir string) {
				mockFS.EXPECT().Stat(dataDir).Return(nil, nil)
			},
			expectError: false,
		},
		{
			name:    "Error - Failed to create directory",
			dataDir: "/data",
			prepareMocks: func(mockFS *mocks.FileSystem, dataDir string) {
				mockFS.EXPECT().Stat(dataDir).Return(nil, os.ErrNotExist)
				mockFS.EXPECT().MkdirAll(dataDir, os.FileMode(0755)).Return(errors.New("mkdir error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := mocks.NewFileSystem(t)
			tt.prepareMocks(mockFS, tt.dataDir)

			storage, err := segment.NewFileSegmentStorage(mockFS, tt.dataDir)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, storage)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, storage)
			}
		})
	}
}

func TestFileSegmentStorage_Create(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	tests := []struct {
		name         string
		id           int
		compressed   bool
		prepareMocks func(mockFS *mocks.FileSystem, dataDir string, id int, compressed bool)
		expectError  bool
	}{
		{
			name:       "Success - Create compressed segment",
			id:         1,
			compressed: true,
			prepareMocks: func(mockFS *mocks.FileSystem, dataDir string, id int, compressed bool) {
				filePath := filepath.Join(dataDir, fmt.Sprintf("segment_%d.wal.gzip", id))
				mockFS.EXPECT().Stat(dataDir).Return(nil, nil).Once()
				mockFS.EXPECT().Create(filePath).Return(&os.File{}, nil).Once()
			},
			expectError: false,
		},
		{
			name:       "Error - Failed to create file",
			id:         1,
			compressed: false,
			prepareMocks: func(mockFS *mocks.FileSystem, dataDir string, id int, compressed bool) {
				filePath := filepath.Join(dataDir, fmt.Sprintf("segment_%d.wal", id))
				mockFS.EXPECT().Stat(dataDir).Return(nil, nil).Once()
				mockFS.EXPECT().Create(filePath).Return(nil, errors.New("create error")).Once()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := mocks.NewFileSystem(t)
			dataDir := "/data"
			tt.prepareMocks(mockFS, dataDir, tt.id, tt.compressed)
			storage, err := segment.NewFileSegmentStorage(mockFS, dataDir)
			require.NoError(t, err)

			segment, err := storage.Create(tt.id, tt.compressed)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, segment)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, segment)
			}
		})
	}
}

func TestFileSegmentStorage_Open(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	tests := []struct {
		name         string
		id           int
		prepareMocks func(mockFS *mocks.FileSystem, dataDir string, id int)
		expectError  bool
	}{
		{
			name: "Success - Open uncompressed segment",
			id:   1,
			prepareMocks: func(mockFS *mocks.FileSystem, dataDir string, id int) {
				filePath := filepath.Join(dataDir, fmt.Sprintf("segment_%d.wal", id))
				mockFS.EXPECT().Stat(dataDir).Return(nil, nil).Once()
				mockFS.EXPECT().Stat(filePath).Return(nil, nil).Once()
				mockFS.EXPECT().Open(filePath).Return(&os.File{}, nil).Once()
				s := mocks.NewSizer(t)
				s.EXPECT().Size().Return(1)
				mockFS.EXPECT().Stat(filePath).Return(s, nil).Once()
			},
			expectError: false,
		},
		{
			name: "Error - Segment not found",
			id:   1,
			prepareMocks: func(mockFS *mocks.FileSystem, dataDir string, id int) {
				filePath := filepath.Join(dataDir, fmt.Sprintf("segment_%d.wal", id))
				mockFS.EXPECT().Stat(dataDir).Return(nil, nil).Once()
				mockFS.EXPECT().Stat(filePath).Return(nil, os.ErrNotExist)
				mockFS.EXPECT().Stat(filePath+".gzip").Return(nil, os.ErrNotExist)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := mocks.NewFileSystem(t)
			dataDir := "/data"
			tt.prepareMocks(mockFS, dataDir, tt.id)
			storage, err := segment.NewFileSegmentStorage(mockFS, dataDir)
			require.NoError(t, err)

			segment, err := storage.Open(tt.id)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, segment)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, segment)
			}
		})
	}
}

func TestFileSegmentStorage_Remove(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	tests := []struct {
		name         string
		id           int
		prepareMocks func(mockFS *mocks.FileSystem, dataDir string, id int)
		expectError  bool
	}{
		{
			name: "Success - Remove segment",
			id:   1,
			prepareMocks: func(mockFS *mocks.FileSystem, dataDir string, id int) {
				filePath := filepath.Join(dataDir, fmt.Sprintf("segment_%d.wal", id))
				mockFS.EXPECT().Stat(dataDir).Return(nil, nil).Once()
				mockFS.EXPECT().Remove(filePath).Return(nil)
			},
			expectError: false,
		},
		{
			name: "Error - Failed to remove segment",
			id:   1,
			prepareMocks: func(mockFS *mocks.FileSystem, dataDir string, id int) {
				filePath := filepath.Join(dataDir, fmt.Sprintf("segment_%d.wal", id))
				mockFS.EXPECT().Stat(dataDir).Return(nil, nil).Once()
				mockFS.EXPECT().Remove(filePath).Return(errors.New("remove error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := mocks.NewFileSystem(t)
			dataDir := "/data"
			tt.prepareMocks(mockFS, dataDir, tt.id)
			storage, err := segment.NewFileSegmentStorage(mockFS, dataDir)
			require.NoError(t, err)

			err = storage.Remove(tt.id)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFileSegmentStorage_List(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	tests := []struct {
		name         string
		prepareMocks func(mockFS *mocks.FileSystem, dataDir string)
		expectError  bool
		expectedIDs  []int
	}{
		{
			name: "Success - List segments",
			prepareMocks: func(mockFS *mocks.FileSystem, dataDir string) {
				mockFS.EXPECT().Stat(dataDir).Return(nil, nil).Once()
				mockEntry1 := mocks.NewDirEntry(t)
				mockEntry1.EXPECT().IsDir().Return(true)
				mockEntry2 := mocks.NewDirEntry(t)
				mockEntry2.EXPECT().IsDir().Return(false)
				mockEntry2.EXPECT().Name().Return("segment_2.wal")
				mockFS.EXPECT().ReadDir(dataDir).Return([]os.DirEntry{mockEntry1, mockEntry2}, nil)
			},
			expectError: false,
			expectedIDs: []int{2},
		},
		{
			name: "Error - Failed to read directory",
			prepareMocks: func(mockFS *mocks.FileSystem, dataDir string) {
				mockFS.EXPECT().Stat(dataDir).Return(nil, nil).Once()
				mockFS.EXPECT().ReadDir(dataDir).Return(nil, errors.New("read dir error"))
			},
			expectError: true,
			expectedIDs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := mocks.NewFileSystem(t)
			dataDir := "/data"
			tt.prepareMocks(mockFS, dataDir)
			storage, err := segment.NewFileSegmentStorage(mockFS, dataDir)
			require.NoError(t, err)

			segments, err := storage.List()
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, segments)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedIDs, segments)
			}
		})
	}
}
