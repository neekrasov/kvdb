package engine_test

import (
	"context"
	"testing"

	"github.com/neekrasov/kvdb/internal/database/storage/engine"
	"github.com/neekrasov/kvdb/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryEngine(t *testing.T) {
	ctx := context.Background()
	logger.MockLogger()

	t.Run("Set and Get", func(t *testing.T) {

		e := engine.New()
		e.Set(ctx, "foo", "bar")
		value, exists := e.Get(ctx, "foo")
		require.True(t, exists)
		assert.Equal(t, "bar", value)
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		e := engine.New(engine.WithPartitionNum(1))
		value, exists := e.Get(ctx, "missing")
		assert.False(t, exists)
		assert.Empty(t, value)
	})

	t.Run("Delete existing key", func(t *testing.T) {
		e := engine.New(engine.WithPartitionNum(1))
		e.Set(ctx, "foo", "bar")
		err := e.Del(ctx, "foo")
		require.NoError(t, err)
		value, exists := e.Get(ctx, "foo")
		assert.False(t, exists)
		assert.Empty(t, value)
	})

	t.Run("Delete non-existent key", func(t *testing.T) {
		e := engine.New(engine.WithPartitionNum(1))
		err := e.Del(ctx, "missing")
		assert.NoError(t, err)
	})
}
