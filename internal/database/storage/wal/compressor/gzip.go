package compressor

import (
	"bytes"
	"compress/gzip"
	"io"
)

// GzipCompressor - provides methods for compressing and decompressing data using the Gzip algorithm
type GzipCompressor struct{}

// Compress - compresses input data ([]byte) using Gzip.
func (g *GzipCompressor) Compress(data []byte) ([]byte, error) {
	if g == nil {
		return data, nil
	}

	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	if _, err := writer.Write(data); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Decompress - decompresses compressed data ([]byte) compressed using Gzip.
func (g *GzipCompressor) Decompress(data []byte) ([]byte, error) {
	if g == nil {
		return data, nil
	}

	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}
