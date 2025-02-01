package conversiontypes

import (
	"unsafe"
)

// UnsafeBytesToString converts a byte slice to a string without allocating new memory.
func UnsafeBytesToString(bytes []byte) string {
	if len(bytes) == 0 {
		return ""
	}

	return unsafe.String(unsafe.SliceData(bytes), len(bytes))
}

// UnsafeStringToBytes converts a string to a byte slice without allocating new memory.
func UnsafeStringToBytes(s string) []byte {
	if s == "" {
		return nil
	}

	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// UnsafeIntToInt64 converts an int to int64 without memory allocation.
func UnsafeIntToInt64(i int) int64 {
	return *(*int64)(unsafe.Pointer(&i))
}
