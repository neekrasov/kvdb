package compressor_test

import (
	"testing"

	"github.com/neekrasov/kvdb/internal/database/storage/wal/compressor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipCompressor_Compress(t *testing.T) {
	g := new(compressor.GzipCompressor)
	data := []byte("test data")

	compressed, err := g.Compress(data)
	require.NoError(t, err, "Compress should not return an error")
	assert.NotEqual(t, data, compressed, "Compressed data should not match original data")

	decompressed, err := g.Decompress(compressed)
	require.NoError(t, err, "Decompress should not return an error")
	assert.Equal(t, data, decompressed, "Decompressed data should match original data")
}

func TestGzipCompressor_Decompress(t *testing.T) {
	g := new(compressor.GzipCompressor)
	data := []byte("test data")

	compressed, err := g.Compress(data)
	require.NoError(t, err, "Compress should not return an error")

	decompressed, err := g.Decompress(compressed)
	require.NoError(t, err, "Decompress should not return an error")
	assert.Equal(t, data, decompressed, "Decompressed data should match original data")
}

func TestGzipCompressor_Decompress_InvalidData(t *testing.T) {
	g := new(compressor.GzipCompressor)
	invalidData := []byte("invalid gzip data")

	_, err := g.Decompress(invalidData)
	assert.Error(t, err, "Decompress should return an error for invalid data")
}

func TestGzipCompressor_NilReceiver(t *testing.T) {
	var g *compressor.GzipCompressor
	data := []byte("test data")

	compressed, err := g.Compress(data)
	require.NoError(t, err, "Compress should not return an error with nil receiver")
	assert.Equal(t, data, compressed, "Compressed data should match original data with nil receiver")

	decompressed, err := g.Decompress(data)
	require.NoError(t, err, "Decompress should not return an error with nil receiver")
	assert.Equal(t, data, decompressed, "Decompressed data should match original data with nil receiver")
}
