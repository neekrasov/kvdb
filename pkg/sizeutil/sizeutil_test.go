package sizeutil_test

import (
	"testing"

	"github.com/neekrasov/kvdb/pkg/sizeutil"
	"github.com/stretchr/testify/assert"
)

func TestParseSize(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input       string
		expected    int
		expectedErr error
	}{
		"valid GB size": {
			input:    "10GB",
			expected: 10 << 30,
		},
		"valid MB size": {
			input:    "5MB",
			expected: 5 << 20,
		},
		"valid KB size": {
			input:    "100KB",
			expected: 100 << 10,
		},
		"valid B size": {
			input:    "200B",
			expected: 200,
		},
		"valid lowercase GB size": {
			input:    "3gb",
			expected: 3 << 30,
		},
		"valid lowercase MB size": {
			input:    "10mb",
			expected: 10 << 20,
		},
		"valid empty size (B)": {
			input:    "0B",
			expected: 0,
		},
		"invalid size with text": {
			input:       "10GBB",
			expected:    0,
			expectedErr: assert.AnError,
		},
		"invalid size with no number": {
			input:       "GB",
			expected:    0,
			expectedErr: assert.AnError,
		},
		"empty string": {
			input:       "",
			expected:    0,
			expectedErr: assert.AnError,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, err := sizeutil.ParseSize(test.input)
			if test.expectedErr != nil {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.expected, result)
		})
	}
}
