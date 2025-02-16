package engine_test

import (
	"testing"

	"github.com/neekrasov/kvdb/internal/database"
	"github.com/neekrasov/kvdb/internal/database/storage/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryEngine(t *testing.T) {
	t.Run("Set and Get", func(t *testing.T) {
		e := engine.New()
		e.Set("foo", "bar")
		value, exists := e.Get("foo")
		require.True(t, exists)
		assert.Equal(t, "bar", value)
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		e := engine.New(engine.WithPartitionNum(1))
		value, exists := e.Get("missing")
		assert.False(t, exists)
		assert.Empty(t, value)
	})

	t.Run("Delete existing key", func(t *testing.T) {
		e := engine.New(engine.WithPartitionNum(1))
		e.Set("foo", "bar")
		err := e.Del("foo")
		require.NoError(t, err)
		value, exists := e.Get("foo")
		assert.False(t, exists)
		assert.Empty(t, value)
	})

	t.Run("Delete non-existent key", func(t *testing.T) {
		e := engine.New(engine.WithPartitionNum(1))
		err := e.Del("missing")
		assert.ErrorIs(t, err, database.ErrKeyNotFound)
	})
}
