package wal_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/neekrasov/kvdb/internal/database/compute"
	"github.com/neekrasov/kvdb/internal/database/storage/wal"
	mocks "github.com/neekrasov/kvdb/internal/mocks/wal"
	"github.com/neekrasov/kvdb/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestWAL_SetAndDel(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	tests := []struct {
		name         string
		prepareMocks func(mockSegmentManager *mocks.SegmentManager)
		operation    func(ctx context.Context, w *wal.WAL) error
		expectError  bool
		cancel       bool
		timeout      time.Duration
	}{
		{
			name: "Success - Set key-value first ctx done",
			prepareMocks: func(mockSegmentManager *mocks.SegmentManager) {
			},
			operation: func(ctx context.Context, w *wal.WAL) error {
				return nil
			},
			expectError: false,
			timeout:     time.Millisecond * 100,
		},

		{
			name: "Success - Set key-value ticker flush",
			prepareMocks: func(mockSegmentManager *mocks.SegmentManager) {
				mockSegmentManager.On("Write", mock.Anything, false).Run(
					func(args mock.Arguments) {
						entries := args.Get(0).([]wal.WriteEntry)
						for _, entry := range entries {
							entry.Set(nil)
						}
					},
				).Return(nil).Once()
			},
			operation: func(ctx context.Context, w *wal.WAL) error {
				return w.Set(ctx, "key1", "value1")
			},
			expectError: false,
			timeout:     time.Millisecond * 10,
		},
		{
			name: "Success - Set key-value flush batches",
			prepareMocks: func(mockSegmentManager *mocks.SegmentManager) {
				mockSegmentManager.On("Write", mock.Anything, false).Run(
					func(args mock.Arguments) {
						entries := args.Get(0).([]wal.WriteEntry)
						for _, entry := range entries {
							entry.Set(nil)
						}
					}).Return(nil).Once()
			},
			operation: func(ctx context.Context, w *wal.WAL) error {
				wg := errgroup.Group{}
				wg.Go(func() error {
					return w.Set(ctx, "key1", "value1")
				})
				wg.Go(func() error {
					return w.Set(ctx, "key2", "value2")
				})
				return wg.Wait()
			},
			expectError: false,
			timeout:     time.Millisecond * 10,
		},
		{
			name: "Success - Delete key",
			prepareMocks: func(mockSegmentManager *mocks.SegmentManager) {
				mockSegmentManager.On("Write", mock.Anything, false).Run(
					func(args mock.Arguments) {
						entries := args.Get(0).([]wal.WriteEntry)
						for _, entry := range entries {
							entry.Set(nil)
						}
					},
				).Return(nil).Once()
			},
			operation: func(ctx context.Context, w *wal.WAL) error {
				return w.Del(ctx, "key1")
			},
			expectError: false,
			timeout:     time.Millisecond,
		},
		{
			name: "Error - Write failed",
			prepareMocks: func(mockSegmentManager *mocks.SegmentManager) {
				mockSegmentManager.On("Write", mock.Anything, false).Run(
					func(args mock.Arguments) {
						time.Sleep(time.Millisecond * 5)
						entries := args.Get(0).([]wal.WriteEntry)
						for _, entry := range entries {
							entry.Set(nil)
						}
					}).Return(errors.New("write error"))
			},
			operation: func(ctx context.Context, w *wal.WAL) error {
				return w.Set(ctx, "key1", "value1")
			},
			expectError: false,
			timeout:     time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSegmentManager := mocks.NewSegmentManager(t)
			tt.prepareMocks(mockSegmentManager)

			w := wal.NewWAL(mockSegmentManager, 2, tt.timeout)
			ctx, cancel := context.WithCancel(context.Background())
			w.Start(ctx)

			err := tt.operation(ctx, w)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			cancel()
			mockSegmentManager.AssertExpectations(t)
		})
	}
}

func TestWAL_Recover(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	tests := []struct {
		name         string
		prepareMocks func(mockSegmentManager *mocks.SegmentManager)
		applyFunc    func(ctx context.Context, entry []wal.LogEntry) error
		expectError  bool
	}{
		{
			name: "Error - error decoding log entry",
			prepareMocks: func(mockSegmentManager *mocks.SegmentManager) {
				mockSegmentManager.On("ForEach", mock.Anything).Run(func(args mock.Arguments) {
					action := args.Get(0).(func(context.Context, []byte) error)
					_ = action(context.Background(), []byte("invalid data"))
				}).Return(nil).Once()
			},
			applyFunc: func(ctx context.Context, entry []wal.LogEntry) error {
				assert.Equal(t, compute.SetCommandID, entry[0].Operation)
				assert.Equal(t, []string{"key1", "value1"}, entry[0].Args)
				return nil
			},
			expectError: false,
		},
		{
			name: "Error - ForEach failed",
			prepareMocks: func(mockSegmentManager *mocks.SegmentManager) {
				mockSegmentManager.On("ForEach", mock.Anything).Return(errors.New("foreach error")).Once()
			},
			applyFunc: func(ctx context.Context, entry []wal.LogEntry) error {
				return nil
			},
			expectError: true,
		},
		{
			name: "Success - Nil apply func",
			prepareMocks: func(mockSegmentManager *mocks.SegmentManager) {
			},
			applyFunc:   nil,
			expectError: false,
		},
		{
			name: "Error - applyFunc returns error",
			prepareMocks: func(mockSegmentManager *mocks.SegmentManager) {
				mockSegmentManager.On("ForEach", mock.Anything).Run(func(args mock.Arguments) {
					action := args.Get(0).(func(context.Context, []byte) error)
					entry := wal.LogEntry{
						Operation: compute.SetCommandID,
						Args:      []string{"key1", "value1"},
					}
					var buffer bytes.Buffer
					if err := entry.Encode(&buffer); err != nil {
						t.Fatalf("failed to encode entry: %v", err)
					}
					_ = action(context.Background(), buffer.Bytes())
				}).Return(nil).Once()
			},
			applyFunc: func(ctx context.Context, entry []wal.LogEntry) error {
				return errors.New("apply func error")
			},
			expectError: false,
		},
		{
			name: "Success - applyFunc returns nil",
			prepareMocks: func(mockSegmentManager *mocks.SegmentManager) {
				mockSegmentManager.On("ForEach", mock.Anything).Run(func(args mock.Arguments) {
					action := args.Get(0).(func(context.Context, []byte) error)
					entry := wal.LogEntry{
						Operation: compute.SetCommandID,
						Args:      []string{"key1", "value1"},
					}
					var buffer bytes.Buffer
					if err := entry.Encode(&buffer); err != nil {
						t.Fatalf("failed to encode entry: %v", err)
					}
					_ = action(context.Background(), buffer.Bytes())
				}).Return(nil).Once()
			},
			applyFunc: func(ctx context.Context, entry []wal.LogEntry) error {
				return nil
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSegmentManager := mocks.NewSegmentManager(t)
			tt.prepareMocks(mockSegmentManager)

			w := wal.NewWAL(mockSegmentManager, 1, time.Second)

			_, err := w.Recover(tt.applyFunc)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockSegmentManager.AssertExpectations(t)
		})
	}
}

func TestWAL_Close(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	mockSegmentManager := mocks.NewSegmentManager(t)
	mockSegmentManager.EXPECT().Close().Return(nil)
	w := wal.NewWAL(mockSegmentManager, 1, time.Second)
	err := w.Close()
	require.NoError(t, err)

	w = nil
	err = w.Close()
	require.NoError(t, err)
}
