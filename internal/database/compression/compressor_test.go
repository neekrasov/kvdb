package compression_test

import (
	"testing"

	"github.com/neekrasov/kvdb/internal/database/compression"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompressors - табличный тест для всех компрессоров
func TestCompressors(t *testing.T) {
	tests := []struct {
		name        string
		compression compression.Compressor
		data        []byte
		invalidData []byte
	}{
		{
			name:        "Bzip2Compressor",
			compression: new(compression.Bzip2Compressor),
			data:        []byte("test data for flate"),
			invalidData: []byte("invalid flate data"),
		},
		{
			name:        "FlateCompressor",
			compression: new(compression.FlateCompressor),
			data:        []byte("test data for flate"),
			invalidData: []byte("invalid flate data"),
		},
		{
			name:        "GzipCompressor",
			compression: new(compression.GzipCompressor),
			data:        []byte("test data for gzip"),
			invalidData: []byte("invalid gzip data"),
		},
		{
			name:        "ZstdCompressor",
			compression: new(compression.ZstdCompressor),
			data:        []byte("test data for zstd"),
			invalidData: []byte("invalid zstd data"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressed, err := tt.compression.Compress(tt.data)
			require.NoError(t, err, "Compress should not return an error")
			assert.NotEqual(t, tt.data, compressed, "Compressed data should not match original data")

			decompressed, err := tt.compression.Decompress(compressed)
			require.NoError(t, err, "Decompress should not return an error")
			assert.Equal(t, tt.data, decompressed, "Decompressed data should match original data")

			_, err = tt.compression.Decompress(tt.invalidData)
			assert.Error(t, err, "Decompress should return an error for invalid data")
		})
	}
}
