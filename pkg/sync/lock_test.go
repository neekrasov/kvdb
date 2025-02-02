package sync_test

import (
	"sync"
	"testing"

	pkgsync "github.com/neekrasov/kvdb/pkg/sync"
	"github.com/stretchr/testify/require"
)

func TestWithLock(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	locked := false

	pkgsync.WithLock(&mu, func() {
		locked = true
	})

	require.True(t, locked)
}

func TestWithLock_NilAction(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	pkgsync.WithLock(&mu, nil)
}
