package storage_test

import (
	"testing"

	"github.com/neekrasov/kvdb/internal/database/models"
	"github.com/neekrasov/kvdb/internal/database/storage"
	mocks "github.com/neekrasov/kvdb/internal/mocks/storage"
	"github.com/neekrasov/kvdb/pkg/gob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNamespaceStorage(t *testing.T) {
	mockEngine := mocks.NewEngine(t)
	nsStorage := storage.NewNamespaceStorage(mockEngine)

	t.Run("Test Exists - exists", func(t *testing.T) {
		namespace := "testNamespace"
		key := storage.MakeKey(models.SystemNamespaceNameSpace, namespace)

		mockEngine.On("Get", key).Return("{}", true).Once()
		assert.True(t, nsStorage.Exists(namespace))
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test Exists - not exists", func(t *testing.T) {
		namespace := "nonexistentNamespace"
		key := storage.MakeKey(models.SystemNamespaceNameSpace, namespace)

		mockEngine.On("Get", key).Return("", false).Once()
		assert.False(t, nsStorage.Exists(namespace))
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test Save - success", func(t *testing.T) {
		namespace := "newNamespace"
		key := storage.MakeKey(models.SystemNamespaceNameSpace, namespace)

		mockEngine.On("Get", key).Return("", false).Once()
		mockEngine.On("Set", key, "{}").Return().Once()

		err := nsStorage.Save(namespace)
		assert.NoError(t, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test Save - already exists", func(t *testing.T) {
		namespace := "existingNamespace"
		key := storage.MakeKey(models.SystemNamespaceNameSpace, namespace)

		mockEngine.On("Get", key).Return("{}", true).Once()

		err := nsStorage.Save(namespace)
		assert.Equal(t, models.ErrNamespaceAlreadyExists, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test Delete - success", func(t *testing.T) {
		namespace := "deleteNamespace"
		key := storage.MakeKey(models.SystemNamespaceNameSpace, namespace)

		mockEngine.On("Get", key).Return("{}", true).Once()
		mockEngine.On("Del", key).Return(nil).Once()

		err := nsStorage.Delete(namespace)
		assert.NoError(t, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test Delete - not found", func(t *testing.T) {
		namespace := "nonexistentNamespace"
		key := storage.MakeKey(models.SystemNamespaceNameSpace, namespace)

		mockEngine.On("Get", key).Return("", false).Once()

		err := nsStorage.Delete(namespace)
		assert.Equal(t, models.ErrNamespaceNotFound, err)
		mockEngine.AssertExpectations(t)
	})

	t.Run("Test Append - success", func(t *testing.T) {
		namespace := "appendNamespace"
		key := models.SystemNamespacesKey

		listBytes, _ := gob.Encode([]string{})

		mockEngine.On("Get", key).Return(string(listBytes), true).Once()
		mockEngine.On("Set", key, mock.Anything).Return().Once()

		namespaces, err := nsStorage.Append(namespace)
		assert.NoError(t, err)
		assert.Contains(t, namespaces, namespace)
		mockEngine.AssertExpectations(t)
	})
}
