package storage_test

import (
	"context"
	"testing"

	"github.com/neekrasov/kvdb/internal/database/storage"
	mocks "github.com/neekrasov/kvdb/internal/mocks/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStorage(t *testing.T) {
	mockEngine := mocks.NewEngine(t)
	mockWAL := mocks.NewWAL(t)

	ctx := context.Background()
	mockWAL.On("Recover", mock.Anything).Return(int64(0), nil)
	store, err := storage.NewStorage(ctx, mockEngine, storage.WithWALOpt(mockWAL))
	require.NoError(t, err)

	t.Run("Set", func(t *testing.T) {
		key, value := "testKey", "testValue"
		mockEngine.On("Set", mock.Anything, key, value).Return().Once()
		mockWAL.On("Set", mock.Anything, key, value).Return(nil).Once()

		err := store.Set(ctx, key, value)
		require.NoError(t, err)
		mockEngine.AssertCalled(t, "Set", mock.Anything, key, value)
	})

	t.Run("Get - Found", func(t *testing.T) {
		key, value := "testKey", "testValue"
		mockEngine.On("Get", mock.Anything, key).Return(value, true).Once()

		result, err := store.Get(ctx, key)

		assert.NoError(t, err)
		assert.Equal(t, value, result)
		mockEngine.AssertCalled(t, "Get", mock.Anything, key)
	})

	t.Run("Get - Not Found", func(t *testing.T) {
		key := "missingKey"
		mockEngine.On("Get", mock.Anything, key).Return("", false).Once()

		result, err := store.Get(ctx, key)

		assert.ErrorIs(t, err, storage.ErrKeyNotFound)
		assert.Empty(t, result)
		mockEngine.AssertCalled(t, "Get", mock.Anything, key)
	})

	t.Run("Del - Success", func(t *testing.T) {
		key := "testKey"
		mockEngine.On("Del", mock.Anything, key).Return(nil).Once()
		mockWAL.On("Del", mock.Anything, key).Return(nil).Once()

		err := store.Del(ctx, key)

		assert.NoError(t, err)
		mockEngine.AssertCalled(t, "Del", mock.Anything, key)
	})

	t.Run("Del - Error", func(t *testing.T) {
		key := "testKey"
		mockEngine.On("Del", mock.Anything, key).Return(storage.ErrKeyNotFound).Once()
		mockWAL.On("Del", mock.Anything, key).Return(nil).Once()

		err := store.Del(ctx, key)

		assert.ErrorIs(t, err, storage.ErrKeyNotFound)
		mockEngine.AssertCalled(t, "Del", mock.Anything, key)
	})
}
