package segment

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/neekrasov/kvdb/internal/database/storage/wal"
	"github.com/neekrasov/kvdb/pkg/logger"
	"go.uber.org/zap"
)

type (
	// Sizer - interface using for mock fileInfo size.
	Sizer interface {
		Size() int64
	}

	// FileSystem - interface used to interact with the file system.
	FileSystem interface {
		Create(name string) (io.ReadWriteCloser, error)
		Open(name string) (io.ReadWriteCloser, error)
		Remove(name string) error
		ReadDir(dirname string) ([]os.DirEntry, error)
		MkdirAll(path string, perm os.FileMode) error
		Stat(name string) (Sizer, error)
	}
)

// FileSegmentStorage - manages segment files on the file system.
type FileSegmentStorage struct {
	fs      FileSystem
	dataDir string
}

// NewFileSegmentStorage - initializes a new FileSegmentStorage.
func NewFileSegmentStorage(fs FileSystem, dataDir string) (*FileSegmentStorage, error) {
	if _, err := fs.Stat(dataDir); os.IsNotExist(err) {
		if err := fs.MkdirAll(dataDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %w", err)
		}
		logger.Debug("created data directory", zap.String("dir", dataDir))
	} else {
		logger.Debug("data directory already exists", zap.String("dir", dataDir))
	}

	return &FileSegmentStorage{dataDir: dataDir, fs: fs}, nil
}

// Create - creates a new segment file.
func (fss *FileSegmentStorage) Create(id int, compressed bool) (wal.Segment, error) {
	filePath := fmt.Sprintf("segment_%d.wal", id)
	if compressed {
		filePath += ".gzip"
	}

	file, err := fss.fs.Create(filepath.Join(fss.dataDir, filePath))
	if err != nil {
		return nil, err
	}
	logger.Debug("created segment", zap.String("filename", filePath), zap.Int("id", id))

	return NewSegment(id, 0, compressed, file), nil
}

// Open - opens an existing segment file.
func (fss *FileSegmentStorage) Open(id int) (wal.Segment, error) {
	path := filepath.Join(fss.dataDir, fmt.Sprintf("segment_%d.wal", id))
	var compressed bool
	if _, err := fss.fs.Stat(path); os.IsNotExist(err) {
		path += ".gzip"
		if _, err := fss.fs.Stat(path); os.IsNotExist(err) {
			return nil, wal.ErrSegmentNotFound
		} else {
			compressed = true
		}
	}

	file, err := fss.fs.Open(path)
	if err != nil {
		return nil, err
	}

	info, err := fss.fs.Stat(path)
	if err != nil {
		return nil, err
	}

	logger.Debug("openned segment",
		zap.String("filename", path),
		zap.Bool("compressed", compressed),
		zap.Int64("size", info.Size()),
		zap.Int("id", id))

	return NewSegment(id, int(info.Size()), compressed, file), nil
}

// Remove - removes a segment file.
func (fss *FileSegmentStorage) Remove(id int) error {
	path := filepath.Join(fss.dataDir, fmt.Sprintf("segment_%d.wal", id))
	logger.Debug("remove segment",
		zap.String("filename", path),
		zap.Int("id", id))
	return fss.fs.Remove(path)
}

// List - lists all segment IDs in the storage.
func (fss *FileSegmentStorage) List() ([]int, error) {
	entries, err := fss.fs.ReadDir(fss.dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list segments: %w", err)
	}

	var segments []int
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		var segmentID int
		_, err := fmt.Sscanf(entry.Name(), "segment_%d.wal", &segmentID)
		if err != nil {
			continue
		}

		segments = append(segments, segmentID)
	}

	sort.Ints(segments)
	logger.Debug("stat segments", zap.Ints("segments", segments))

	return segments, nil
}
