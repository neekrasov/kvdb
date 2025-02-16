package segment_test

import (
	"bytes"
	"testing"

	"github.com/neekrasov/kvdb/internal/database/storage/wal/segment"
	"github.com/stretchr/testify/assert"
)

type BufferCloser struct {
	*bytes.Buffer
}

func (b *BufferCloser) Close() error {
	return nil
}

func NewBufferCloser() *BufferCloser {
	return &BufferCloser{Buffer: &bytes.Buffer{}}
}

func TestNewSegment(t *testing.T) {
	buf := NewBufferCloser()
	segment := segment.NewSegment(1, 100, true, buf)

	// Проверяем, что поля структуры инициализированы правильно
	assert.Equal(t, 1, segment.ID())
	assert.Equal(t, 100, segment.Size())
	assert.Equal(t, true, segment.Compressed())
}

func TestSegment_Write_Success(t *testing.T) {
	buf := NewBufferCloser()
	segment := segment.NewSegment(1, 100, true, buf)

	data := []byte("test data")
	n, err := segment.Write(data)

	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, 100+len(data), segment.Size())
	assert.Equal(t, "test data", buf.String())
}

func TestSegment_Read_Success(t *testing.T) {
	buf := NewBufferCloser()
	buf.WriteString("test data")
	segment := segment.NewSegment(1, 100, true, buf)

	buffer := make([]byte, 10)
	n, err := segment.Read(buffer)

	assert.NoError(t, err)
	assert.Equal(t, 9, n)
	assert.Equal(t, "test data", string(buffer[:n]))
}

func TestSegment_Read_EOF(t *testing.T) {
	buf := NewBufferCloser()
	segment := segment.NewSegment(1, 100, true, buf)

	buffer := make([]byte, 10)
	n, err := segment.Read(buffer)

	assert.Error(t, err)
	assert.Equal(t, 0, n)
}

func TestSegment_Close(t *testing.T) {
	buf := NewBufferCloser()
	segment := segment.NewSegment(1, 100, true, buf)

	err := segment.Close()

	assert.NoError(t, err)
}

func TestSegment_Size(t *testing.T) {
	buf := NewBufferCloser()
	segment := segment.NewSegment(1, 100, true, buf)

	assert.Equal(t, 100, segment.Size())
}
