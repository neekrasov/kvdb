package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/neekrasov/kvdb/internal/database/storage"
	mocks "github.com/neekrasov/kvdb/internal/mocks/storage"
	"github.com/neekrasov/kvdb/pkg/ctxutil"
	"github.com/neekrasov/kvdb/pkg/logger"
	pkgsync "github.com/neekrasov/kvdb/pkg/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStorage(t *testing.T) {
	t.Parallel()

	mockEngine := mocks.NewEngine(t)
	mockWAL := mocks.NewWAL(t)

	ctx := context.Background()
	mockWAL.On("Recover", mock.Anything).Return(int64(0), nil)
	store, err := storage.NewStorage(ctx, mockEngine, storage.WithWALOpt(mockWAL))
	require.NoError(t, err)

	t.Run("Set", func(t *testing.T) {
		key, value := "testKey", "testValue"
		mockEngine.On("Set", mock.Anything, key, value, int64(0)).Return().Once()
		mockWAL.On("Set", mock.Anything, key, value).Return(nil).Once()

		err := store.Set(ctx, key, value)
		require.NoError(t, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Get - Found", func(t *testing.T) {
		key, value := "testKey", "testValue"
		mockEngine.On("Get", mock.Anything, key).Return(value, true).Once()

		result, err := store.Get(ctx, key)

		assert.NoError(t, err)
		assert.Equal(t, value, result)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Get - Not Found", func(t *testing.T) {
		key := "missingKey"
		mockEngine.On("Get", mock.Anything, key).Return("", false).Once()

		result, err := store.Get(ctx, key)

		assert.ErrorIs(t, err, storage.ErrKeyNotFound)
		assert.Empty(t, result)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Del - Success", func(t *testing.T) {
		key := "testKey"
		mockEngine.On("Del", mock.Anything, key).Return(nil).Once()
		mockWAL.On("Del", mock.Anything, key).Return(nil).Once()

		err := store.Del(ctx, key)

		assert.NoError(t, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Del - Error", func(t *testing.T) {
		key := "testKey"
		mockEngine.On("Del", mock.Anything, key).Return(storage.ErrKeyNotFound).Once()
		mockWAL.On("Del", mock.Anything, key).Return(nil).Once()

		err := store.Del(ctx, key)

		assert.ErrorIs(t, err, storage.ErrKeyNotFound)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Watch", func(t *testing.T) {
		key := "testKey"
		mockEngine.On("Watch", mock.Anything, key).Return(pkgsync.NewFuture[string]()).Once()

		future := store.Watch(ctx, key)
		assert.NotNil(t, future)
	})

	t.Run("Set with invalid TTL", func(t *testing.T) {
		key, value := "testKey", "testValue"
		ttlCtx := ctxutil.InjectTTL(ctx, "1klmn")

		err := store.Set(ttlCtx, key, value)
		require.Error(t, err)
		require.ErrorContains(t, err, "invalid format fo ttl:")
		mockEngine.AssertExpectations(t)
	})

	t.Run("Set with TTL", func(t *testing.T) {
		key, value := "testKey", "testValue"
		mockEngine.On("Set", mock.Anything, key, value, mock.Anything).Return().Once()
		mockWAL.On("Set", mock.Anything, key, value).Return(nil).Once()

		ttlCtx := ctxutil.InjectTTL(ctx, "10m")

		err := store.Set(ttlCtx, key, value)
		require.NoError(t, err)
		mockEngine.AssertExpectations(t)
	})

}

func TestStorageCleanupBackground(t *testing.T) {
	logger.MockLogger()

	mockEngine := mocks.NewEngine(t)
	mockWAL := mocks.NewWAL(t)

	mockWAL.On("Recover", mock.Anything).Return(int64(0), nil)
	mockWAL.On("Flush", mock.Anything).Return(nil)

	expiredKeys := []string{"expiredKey1", "expiredKey2", "expiredKey3"}
	mockEngine.On("ForEachExpired", mock.Anything).Run(func(args mock.Arguments) {
		action := args.Get(0).(func(string))
		for _, key := range expiredKeys {
			action(key)
		}
	}).Return().Once()
	mockEngine.On("ForEachExpired", mock.Anything).Return()

	for _, key := range expiredKeys {
		mockEngine.On("Del", mock.Anything, key).Return(nil).Once()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cleanupPeriod := 100 * time.Millisecond
	cleanupBatchSize := 2

	_, err := storage.NewStorage(
		ctx,
		mockEngine,
		storage.WithWALOpt(mockWAL),
		storage.WithCleanupPeriod(cleanupPeriod),
		storage.WithCleanupBatchSize(cleanupBatchSize),
	)
	require.NoError(t, err)

	time.Sleep(cleanupPeriod * 2)

	mockEngine.AssertExpectations(t)
	mockWAL.AssertExpectations(t)

	mockEngine.AssertNumberOfCalls(t, "Del", 3)
}
