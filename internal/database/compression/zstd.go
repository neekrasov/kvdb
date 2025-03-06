package compression

import (
	"bytes"
	"io"

	"github.com/klauspost/compress/zstd"
)

// ZstdCompressor - реализация сжатия и распаковки с использованием Zstandard
type ZstdCompressor struct{}

func (z *ZstdCompressor) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := zstd.NewWriter(&buf)
	if err != nil {
		return nil, err
	}
	if _, err := writer.Write(data); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (z *ZstdCompressor) Decompress(data []byte) ([]byte, error) {
	reader, err := zstd.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}
