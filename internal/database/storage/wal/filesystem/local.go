package filesystem

import (
	"io"
	"os"

	"github.com/neekrasov/kvdb/internal/database/storage/wal/segment"
)

// LocalFileSystem implements a file system interface using the local operating system's file system.
type LocalFileSystem struct{}

// Create creates a new file with the specified name and returns a ReadWriteCloser for it.
func (fs *LocalFileSystem) Create(name string) (io.ReadWriteCloser, error) {
	return os.Create(name)
}

// Open opens an existing file with the specified name and returns a ReadWriteCloser for it.
func (fs *LocalFileSystem) Open(name string) (io.ReadWriteCloser, error) {
	return os.Open(name)
}

// Remove deletes the file or directory with the specified name.
func (fs *LocalFileSystem) Remove(name string) error {
	return os.Remove(name)
}

// ReadDir reads the directory named by dirname and returns a list of directory entries.
func (fs *LocalFileSystem) ReadDir(dirname string) ([]os.DirEntry, error) {
	return os.ReadDir(dirname)
}

// MkdirAll creates a directory named path, along with any necessary parents, with the specified permissions.
func (fs *LocalFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// Stat returns the FileInfo structure describing the file or directory named by name.
func (fs *LocalFileSystem) Stat(name string) (segment.Sizer, error) {
	return os.Stat(name)
}
