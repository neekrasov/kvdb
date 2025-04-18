package identity_test

import (
	"context"
	"testing"

	"github.com/neekrasov/kvdb/internal/database/identity"
	"github.com/neekrasov/kvdb/internal/database/identity/models"
	"github.com/neekrasov/kvdb/internal/database/storage"
	mocks "github.com/neekrasov/kvdb/internal/mocks/database"
	"github.com/neekrasov/kvdb/pkg/gob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNamespaceStorage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockStorage := mocks.NewStorage(t)
	nsStorage := identity.NewNamespaceStorage(mockStorage)

	t.Run("Test Exists - exists", func(t *testing.T) {
		namespace := "testNamespace"
		key := storage.MakeKey(models.SystemNamespaceNameSpace, namespace)

		mockStorage.On("Get", mock.Anything, key).Return("{}", nil).Once()
		assert.True(t, nsStorage.Exists(ctx, namespace))
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test Exists - not exists", func(t *testing.T) {
		namespace := "nonexistentNamespace"
		key := storage.MakeKey(models.SystemNamespaceNameSpace, namespace)

		mockStorage.On("Get", mock.Anything, key).Return("", storage.ErrKeyNotFound).Once()
		assert.False(t, nsStorage.Exists(ctx, namespace))
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test Save - success", func(t *testing.T) {
		namespace := "newNamespace"
		key := storage.MakeKey(models.SystemNamespaceNameSpace, namespace)

		mockStorage.On("Get", mock.Anything, key).Return("", storage.ErrKeyNotFound).Once()
		mockStorage.On("Set", mock.Anything, key, "").Return(nil).Once()

		err := nsStorage.Save(ctx, namespace)
		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test Save - already exists", func(t *testing.T) {
		namespace := "existingNamespace"
		key := storage.MakeKey(models.SystemNamespaceNameSpace, namespace)

		mockStorage.On("Get", mock.Anything, key).Return("{}", nil).Once()

		err := nsStorage.Save(ctx, namespace)
		assert.Equal(t, identity.ErrNamespaceAlreadyExists, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test Delete - success", func(t *testing.T) {
		namespace := "deleteNamespace"
		key := storage.MakeKey(models.SystemNamespaceNameSpace, namespace)

		mockStorage.On("Get", mock.Anything, key).Return("{}", nil).Once()
		mockStorage.On("Del", mock.Anything, key).Return(nil).Once()

		err := nsStorage.Delete(ctx, namespace)
		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test Delete - not found", func(t *testing.T) {
		namespace := "nonexistentNamespace"
		key := storage.MakeKey(models.SystemNamespaceNameSpace, namespace)

		mockStorage.On("Get", mock.Anything, key).Return("", storage.ErrKeyNotFound).Once()

		err := nsStorage.Delete(ctx, namespace)
		assert.Equal(t, identity.ErrNamespaceNotFound, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Test Append - success", func(t *testing.T) {
		namespace := "appendNamespace"
		key := models.SystemNamespacesKey

		listBytes, _ := gob.Encode([]string{})

		mockStorage.On("Get", mock.Anything, key).Return(string(listBytes), nil).Once()
		mockStorage.On("Set", mock.Anything, key, mock.Anything).Return(nil).Once()

		namespaces, err := nsStorage.Append(ctx, namespace)
		assert.NoError(t, err)
		assert.Contains(t, namespaces, namespace)
		mockStorage.AssertExpectations(t)
	})
}
