package storage_test

import (
	"testing"

	"github.com/neekrasov/kvdb/internal/database"
	"github.com/neekrasov/kvdb/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryEngine(t *testing.T) {
	t.Run("Set and Get", func(t *testing.T) {
		e := storage.NewInMemoryEngine()
		e.Set("foo", "bar")
		value, err := e.Get("foo")
		require.NoError(t, err)
		assert.Equal(t, "bar", value)
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		e := storage.NewInMemoryEngine()
		value, err := e.Get("missing")
		assert.ErrorIs(t, err, database.ErrKeyNotFound)
		assert.Empty(t, value)
	})

	t.Run("Delete existing key", func(t *testing.T) {
		e := storage.NewInMemoryEngine()
		e.Set("foo", "bar")
		err := e.Del("foo")
		require.NoError(t, err)
		_, err = e.Get("foo")
		assert.ErrorIs(t, err, database.ErrKeyNotFound)
	})

	t.Run("Delete non-existent key", func(t *testing.T) {
		e := storage.NewInMemoryEngine()
		err := e.Del("missing")
		assert.ErrorIs(t, err, database.ErrKeyNotFound)
	})
}
