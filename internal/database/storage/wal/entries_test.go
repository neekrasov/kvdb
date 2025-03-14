package wal_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/neekrasov/kvdb/internal/database/compute"
	"github.com/neekrasov/kvdb/internal/database/storage/wal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogEntryEncodeDecode(t *testing.T) {
	t.Parallel()

	entry := wal.LogEntry{
		Operation: compute.CommandID(1),
		Args:      []string{"arg1", "arg2"},
	}

	var buf bytes.Buffer
	err := entry.Encode(&buf)
	require.NoError(t, err, "Encode should not return an error")

	var decodedEntry wal.LogEntry
	err = decodedEntry.Decode(&buf)
	require.NoError(t, err, "Decode should not return an error")

	assert.Equal(t, entry.Operation, decodedEntry.Operation, "Operation should match")
	assert.Equal(t, entry.Args, decodedEntry.Args, "Args should match")
}

func TestWriteEntry(t *testing.T) {
	t.Parallel()

	op := compute.CommandID(1)
	args := []string{"arg1", "arg2"}
	entry := wal.NewWriteEntry(0, op, args)

	logEntry := entry.Log()
	assert.Equal(t, op, logEntry.Operation, "Operation should match")
	assert.Equal(t, args, logEntry.Args, "Args should match")

	testErr := errors.New("test error")
	go entry.Set(testErr)

	err := entry.Get()
	assert.ErrorIs(t, err, testErr, "Get should return the set error")
}
