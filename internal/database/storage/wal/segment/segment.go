package segment

import (
	"io"
)

// Segment - represents a single segment file.
type Segment struct {
	compressed bool
	id         int
	file       io.ReadWriteCloser
	size       int
}

// NewSegment - initializes a new Segment.
func NewSegment(id, size int, compressed bool, file io.ReadWriteCloser) *Segment {
	return &Segment{
		file:       file,
		size:       size,
		id:         id,
		compressed: compressed,
	}
}

// Write - writes data to the segment file.
func (s *Segment) Write(data []byte) (int, error) {
	n, err := s.file.Write(data)
	if err != nil {
		return 0, err
	}
	s.size += n

	return n, nil
}

// Read - reads data from the segment file.
func (s *Segment) Read(p []byte) (n int, err error) {
	return s.file.Read(p)
}

// Close - closes the segment file.
func (s *Segment) Close() error {
	return s.file.Close()
}

// Size - returns the size of the segment.
func (s *Segment) Size() int {
	return s.size
}

// ID - returns the ID of the segment.
func (s *Segment) ID() int {
	return s.id
}

// Compressed - returns whether the segment is compressed.
func (s *Segment) Compressed() bool {
	return s.compressed
}
