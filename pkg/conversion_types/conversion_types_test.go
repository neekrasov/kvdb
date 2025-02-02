package conversiontypes_test

import (
	"testing"
	"unsafe"

	conversiontypes "github.com/neekrasov/kvdb/pkg/conversion_types"
)

func TestUnsafeBytesToString(t *testing.T) {
	t.Parallel()

	b := []byte("hello")
	s := conversiontypes.UnsafeBytesToString(b)

	// Check if conversion result is correct
	if s != "hello" {
		t.Errorf("expected 'hello', got '%s'", s)
	}

	// Ensure no memory allocation (pointers should match)
	if unsafe.SliceData(b) != unsafe.StringData(s) {
		t.Errorf("expected memory addresses to match, but they differ")
	}

	// Check empty slice case
	if conversiontypes.UnsafeBytesToString(nil) != "" {
		t.Errorf("expected empty string for nil input")
	}
}

func TestUnsafeStringToBytes(t *testing.T) {
	t.Parallel()

	s := "hello"
	b := conversiontypes.UnsafeStringToBytes(s)

	// Check if conversion result is correct
	if string(b) != "hello" {
		t.Errorf("expected 'hello', got '%s'", string(b))
	}

	// Ensure no memory allocation (pointers should match)
	if unsafe.StringData(s) != unsafe.SliceData(b) {
		t.Errorf("expected memory addresses to match, but they differ")
	}

	// Check empty string case
	if conversiontypes.UnsafeStringToBytes("") != nil {
		t.Errorf("expected nil slice for empty string input")
	}
}

func TestUnsafeIntToInt64(t *testing.T) {
	t.Parallel()

	i := 42
	i64 := conversiontypes.UnsafeIntToInt64(i)

	// Check if conversion preserves value
	if int(i64) != i {
		t.Errorf("expected %d, got %d", i, i64)
	}
}
