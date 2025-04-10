package engine_test

import (
	"context"
	"testing"
	"time"

	"github.com/neekrasov/kvdb/internal/database/storage/engine"
	"github.com/neekrasov/kvdb/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryEngine(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger.MockLogger()

	t.Run("Set and Get", func(t *testing.T) {
		e := engine.New()
		e.Set(ctx, "foo", "bar", 0)
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
		e.Set(ctx, "foo", "bar", 0)
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

	t.Run("Watch", func(t *testing.T) {
		e := engine.New()
		key, value := "test_key", "test_value"
		future := e.Watch(ctx, key)
		time.Sleep(time.Millisecond)
		go func() {
			e.Set(ctx, key, value, 0)
		}()

		actual := future.Get()

		assert.Equal(t, value, actual)
	})

	t.Run("Watch multiple", func(t *testing.T) {
		e := engine.New()
		key, value := "test_key", "test_value"
		future1 := e.Watch(ctx, key)
		future2 := e.Watch(ctx, key)
		time.Sleep(time.Millisecond)
		go func() {
			e.Set(ctx, key, value, 0)
		}()

		actual1 := future1.Get()
		actual2 := future2.Get()

		assert.Equal(t, value, actual1)
		assert.Equal(t, actual1, actual2)
	})

	t.Run("Watch cancel", func(t *testing.T) {
		e := engine.New()
		key, value := "test_key", "test_value"
		e.Set(ctx, key, value, 0)

		ctx, cancel := context.WithCancel(ctx)
		future := e.Watch(ctx, key)
		cancel()

		actual := future.Get()

		assert.Equal(t, value, actual)
	})
}
