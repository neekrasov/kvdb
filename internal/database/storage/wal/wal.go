package wal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/neekrasov/kvdb/internal/database/command"
	pkgSync "github.com/neekrasov/kvdb/pkg/sync"

	"github.com/neekrasov/kvdb/pkg/logger"
	"go.uber.org/zap"
)

// SegmentManager - interface for managing segments.
type SegmentManager interface {
	// Write - writes entries to the current segment.
	Write(data []WriteEntry) error
	// ForEach - iterates through all segments.
	ForEach(action func([]byte) error) error
	// Close - closes the current segment.
	Close() error
}

// WAL - Write-Ahead Log implementation.
type WAL struct {
	segmentManager SegmentManager
	batchSize      int
	flushTimeout   time.Duration
	batches        chan struct{}

	mu    sync.Mutex
	batch []WriteEntry
}

// NewWAL - initializes and returns a new WAL.
func NewWAL(segmentManager SegmentManager, batchSize int, flushTimeout time.Duration) *WAL {
	return &WAL{
		segmentManager: segmentManager,
		batchSize:      batchSize,
		flushTimeout:   flushTimeout,
		batch:          make([]WriteEntry, 0, batchSize),
		batches:        make(chan struct{}),
	}
}

// Start - starts the WAL background flush process.
func (w *WAL) Start(ctx context.Context) {
	baseErrMsg := "failed to flush batch"
	go func() {
		ticker := time.NewTicker(w.flushTimeout)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				if err := w.flushLocked(); err != nil {
					logger.Debug(baseErrMsg, zap.Error(err))
				}
				return
			default:
			}

			select {
			case <-ctx.Done():
				if err := w.flushLocked(); err != nil {
					logger.Debug(baseErrMsg, zap.Error(err))
				}
				return
			case <-w.batches:
				if err := w.flushLocked(); err != nil {
					logger.Debug(baseErrMsg, zap.Error(err))
				}
				ticker.Reset(w.flushTimeout)
			case <-ticker.C:
				if err := w.flushLocked(); err != nil {
					logger.Debug(baseErrMsg, zap.Error(err))
				}
			}
		}
	}()
}

// Set - push a set operation to the WAL.
func (w *WAL) Set(key, value string) error {
	return w.push(command.SetCommandID, []string{key, value})
}

// Delete - push a delete operation to the WAL.
func (w *WAL) Del(key string) error {
	return w.push(command.DelCommandID, []string{key})
}

// push - pushes a log entry to the batch.
func (w *WAL) push(op command.CommandID, args []string) error {
	if w == nil {
		return nil
	}

	logger.Debug(
		"pushed log entry to wal",
		zap.Int("operation", int(op)),
		zap.Strings("args", args),
	)

	we := NewWriteEntry(op, args)
	pkgSync.WithLock(&w.mu, func() {
		w.batch = append(w.batch, we)
		if len(w.batch) >= w.batchSize {
			w.batches <- struct{}{}
		}
	})

	return we.future.Get()
}

// flushLocked - flushes the current batch to the segment.
func (w *WAL) flushLocked() error {
	if len(w.batch) != 0 {
		if err := w.segmentManager.Write(w.batch); err != nil {
			return fmt.Errorf("failed to write to segment: %w", err)
		}

		logger.Debug("flush segments")
		w.batch = nil
	}

	return nil
}

// Recover - recovers the state from the WAL.
func (w *WAL) Recover(applyFunc func(entry LogEntry) error) error {
	if w == nil || applyFunc == nil {
		return nil
	}

	logger.Debug("start recovering segments")
	err := w.segmentManager.ForEach(
		func(b []byte) error {
			buffer := bytes.NewBuffer(b)
			for buffer.Len() > 0 {
				var entry LogEntry
				if err := entry.Decode(buffer); err != nil && err != io.EOF {
					return fmt.Errorf("error gob decoding: %w", err)
				}

				if err := applyFunc(entry); err != nil {
					return fmt.Errorf(
						"error while applying entry cmd id %d with args %v: %w",
						entry.Operation, entry.Args, err,
					)
				}
			}

			return nil
		})
	if err != nil {
		return fmt.Errorf("execute action for recover failed: %w", err)
	}

	return nil
}

// Close - closes the WAL.
func (w *WAL) Close() error {
	if w == nil {
		return nil
	}

	return w.segmentManager.Close()
}
