package storage_test

import (
	"testing"

	"github.com/neekrasov/kvdb/internal/database"
	"github.com/neekrasov/kvdb/internal/database/storage"
	"github.com/neekrasov/kvdb/internal/database/storage/models"
	mocks "github.com/neekrasov/kvdb/internal/mocks/storage"
	"github.com/stretchr/testify/assert"
)

func TestStorage(t *testing.T) {
	mockEngine := mocks.NewEngine(t)
	store := storage.NewStorage(mockEngine)

	t.Run("Set", func(t *testing.T) {
		key, value := "testKey", "testValue"
		mockEngine.On("Set", key, value).Return().Once()

		store.Set(key, value)
		mockEngine.AssertCalled(t, "Set", key, value)
	})

	t.Run("Get - Found", func(t *testing.T) {
		key, value := "testKey", "testValue"
		mockEngine.On("Get", key).Return(value, true).Once()

		result, err := store.Get(key)

		assert.NoError(t, err)
		assert.Equal(t, value, result)
		mockEngine.AssertCalled(t, "Get", key)
	})

	t.Run("Get - Not Found", func(t *testing.T) {
		key := "missingKey"
		mockEngine.On("Get", key).Return("", false).Once()

		result, err := store.Get(key)

		assert.ErrorIs(t, err, models.ErrKeyNotFound)
		assert.Empty(t, result)
		mockEngine.AssertCalled(t, "Get", key)
	})

	t.Run("Del - Success", func(t *testing.T) {
		key := "testKey"
		mockEngine.On("Del", key).Return(nil).Once()

		err := store.Del(key)

		assert.NoError(t, err)
		mockEngine.AssertCalled(t, "Del", key)
	})

	t.Run("Del - Error", func(t *testing.T) {
		key := "testKey"
		mockEngine.On("Del", key).Return(database.ErrKeyNotFound).Once()

		err := store.Del(key)

		assert.ErrorIs(t, err, database.ErrKeyNotFound)
		mockEngine.AssertCalled(t, "Del", key)
	})
}
