package wal

import (
	"encoding/gob"
	"fmt"
	"io"

	"github.com/neekrasov/kvdb/internal/database/compute"
	"github.com/neekrasov/kvdb/pkg/sync"
)

// LogEntry - represents a single log entry in the Write-Ahead Log (WAL).
// It is the minimal unit that is written to the WAL.
type LogEntry struct {
	// LSN (Log Sequence Number) is a unique identifier for the log entry.
	// It is used to ensure the order of log entries and for recovery purposes.
	LSN int64
	// Operation represents the type of operation being logged.
	Operation compute.CommandID
	// Args contains the arguments or parameters associated with the operation.
	Args []string
}

// Encode - encodes a LogEntry.
func (e LogEntry) Encode(w io.Writer) error {
	if err := gob.NewEncoder(w).Encode(e); err != nil {
		return fmt.Errorf("encode failed: %w", err)
	}

	return nil
}

// Decode - decodes a LogEntry.
func (e *LogEntry) Decode(r io.Reader) error {
	if err := gob.NewDecoder(r).Decode(e); err != nil {
		return fmt.Errorf("decode failed: %w", err)
	}

	return nil
}

// WriteEntry - represents a write entry with a future.
type WriteEntry struct {
	log    LogEntry
	future sync.FutureError
}

// NewWriteEntry - creates a new WriteEntry.
func NewWriteEntry(lsn int64, op compute.CommandID, args []string) WriteEntry {
	return WriteEntry{
		log: LogEntry{
			LSN:       lsn,
			Operation: op,
			Args:      args,
		},
		future: sync.NewFuture[error](),
	}
}

// Set - sets the error for the WriteEntry.
func (l *WriteEntry) Set(err error) {
	l.future.Set(err)
}

// Get - gets the error from the WriteEntry.
func (l *WriteEntry) Get() error {
	return l.future.Get()
}

// Log - returns the LogEntry.
func (l *WriteEntry) Log() LogEntry {
	return l.log
}
