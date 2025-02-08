package gob

import (
	"bytes"
	"encoding/gob"
)

// Encode - Helper function to encode data using Gob.
func Encode(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// sDecode - Helper function to decode data using Gob.
func Decode(data []byte, target interface{}) error {
	buf := bytes.NewBuffer(data)
	return gob.NewDecoder(buf).Decode(target)
}
