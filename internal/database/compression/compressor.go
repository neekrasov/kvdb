package compression

import "errors"

// Compressor - interface for data compression and decompression.
type Compressor interface {
	// Compress - compresses input data ([]byte) using some compress algorithm.
	Compress(data []byte) ([]byte, error)
	// Decompress - decompresses compressed data ([]byte) compressed using some compress algorithm.
	Decompress(data []byte) ([]byte, error)
}

// CompressionType - type of compression.
type CompressionType string

const (
	Gzip  CompressionType = "gzip"
	Zstd  CompressionType = "zstd"
	Bzip2 CompressionType = "bzip2"
	Flate CompressionType = "flate"
)

// New - creates a new compression depends compression type.
func New(ct CompressionType) (Compressor, error) {
	switch ct {
	case Gzip:
		return new(GzipCompressor), nil
	case Zstd:
		return new(ZstdCompressor), nil
	case Bzip2:
		return new(Bzip2Compressor), nil
	case Flate:
		return new(FlateCompressor), nil
	}

	return nil, errors.New("unsuported compression type")
}
