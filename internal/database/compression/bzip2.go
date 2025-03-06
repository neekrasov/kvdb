package compression

import (
	"bytes"
	"compress/bzip2"
	"io"

	libzip "github.com/dsnet/compress/bzip2"
)

type Bzip2Compressor struct{}

func (b *Bzip2Compressor) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	config := libzip.WriterConfig{
		Level: libzip.BestCompression,
	}

	writer, err := libzip.NewWriter(&buf, &config)
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

func (b *Bzip2Compressor) Decompress(data []byte) ([]byte, error) {
	return io.ReadAll(bzip2.NewReader(bytes.NewReader(data)))
}
