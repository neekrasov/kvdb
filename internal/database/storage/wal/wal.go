package wal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"sync"
	"time"

	"github.com/neekrasov/kvdb/internal/database/compute"
	"github.com/neekrasov/kvdb/pkg/ctxutil"
	pkgsync "github.com/neekrasov/kvdb/pkg/sync"

	"github.com/neekrasov/kvdb/pkg/logger"
	"go.uber.org/zap"
)

// SegmentManager - interface for managing segments.
type SegmentManager interface {
	// Write - writes entries to the current segment.
	Write(entries []WriteEntry, nolock bool) error
	// ForEach - iterates through all segments.
	ForEach(action func(ctx context.Context, b []byte) error) error
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
				if err := w.flush(); err != nil {
					logger.Debug(baseErrMsg, zap.Error(err))
				}
				return
			default:
			}

			select {
			case <-ctx.Done():
				if err := w.flush(); err != nil {
					logger.Debug(baseErrMsg, zap.Error(err))
				}
				return
			case <-w.batches:
				if err := w.flush(); err != nil {
					logger.Debug(baseErrMsg, zap.Error(err))
				}
				ticker.Reset(w.flushTimeout)
			case <-ticker.C:
				if err := w.flush(); err != nil {
					logger.Debug(baseErrMsg, zap.Error(err))
				}
			}
		}
	}()
}

// Set - push a set operation to the WAL.
func (w *WAL) Set(ctx context.Context, key, value string) error {
	if w == nil {
		return nil
	}

	return w.push(ctx, compute.SetCommandID, []string{key, value})
}

// Delete - push a delete operation to the WAL.
func (w *WAL) Del(ctx context.Context, key string) error {
	if w == nil {
		return nil
	}

	return w.push(ctx, compute.DelCommandID, []string{key})
}

// push - pushes a log entry to the batch.
func (w *WAL) push(ctx context.Context, op compute.CommandID, args []string) error {
	if w == nil {
		return nil
	}

	txID := ctxutil.ExtractTxID(ctx)
	logger.Debug(
		"pushed log entry to wal",
		zap.Int("operation", int(op)),
		zap.Strings("args", args),
		zap.Int64("tx", txID),
	)

	entry := NewWriteEntry(txID, op, args)
	pkgsync.WithLock(&w.mu, func() {
		w.batch = append(w.batch, entry)
		if len(w.batch) >= w.batchSize {
			w.batches <- struct{}{}
		}
	})

	return entry.future.Get()
}

func (w *WAL) Flush(batch []WriteEntry) error {
	if err := w.segmentManager.Write(batch, true); err != nil {
		return fmt.Errorf("failed to write to segment: %w", err)
	}

	logger.Debug("force flush segments")
	return nil
}

// flush - flushes the current batch to the segment.
func (w *WAL) flush() error {
	if len(w.batch) != 0 {
		if err := w.segmentManager.Write(w.batch, false); err != nil {
			return fmt.Errorf("failed to write to segment: %w", err)
		}

		logger.Debug("flush segments")
		w.batch = nil
	}

	return nil
}

// Recover - recovers the state from the WAL.
func (w *WAL) Recover(applyFunc func(ctx context.Context, entry []LogEntry) error) (int64, error) {
	if w == nil || applyFunc == nil {
		return 0, nil
	}

	var lastLSN int64
	logger.Debug("start recovering segments")
	err := w.segmentManager.ForEach(
		func(ctx context.Context, b []byte) error {
			var entries []LogEntry

			buffer := bytes.NewBuffer(b)
			for buffer.Len() > 0 {
				var entry LogEntry
				if err := entry.Decode(buffer); err != nil && err != io.EOF {
					return fmt.Errorf("error gob decoding: %w", err)
				}
				entries = append(entries, entry)
			}

			sort.Slice(entries, func(i, j int) bool {
				return entries[i].LSN < entries[j].LSN
			})

			if err := applyFunc(ctx, entries); err != nil {
				return fmt.Errorf("failed to epply entries: %w", err)
			}

			if len(entries) > 0 {
				lastLSN = entries[len(entries)-1].LSN
			}
			return nil
		})
	if err != nil {
		return 0, fmt.Errorf("execute action for recover failed: %w", err)
	}

	return lastLSN, nil
}

// Close - closes the WAL.
func (w *WAL) Close() error {
	if w == nil {
		return nil
	}

	return w.segmentManager.Close()
}
