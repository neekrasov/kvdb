package ctxutil_test

import (
	"context"
	"testing"

	"github.com/neekrasov/kvdb/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInjectAndExtractTxID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	txID := int64(12345)
	ctx = ctxutil.InjectTxID(ctx, txID)

	extractedTxID := ctxutil.ExtractTxID(ctx)
	assert.Equal(t, txID, extractedTxID, "Expected TxID to be %d, got %d", txID, extractedTxID)
}

func TestExtractTxID_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	extractedTxID := ctxutil.ExtractTxID(ctx)
	assert.Equal(t, int64(0), extractedTxID, "Expected TxID to be 0, got %d", extractedTxID)
}

func TestInjectAndExtractSessionID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	sessionID := "abc123"
	ctx = ctxutil.InjectSessionID(ctx, sessionID)

	extractedSessionID := ctxutil.ExtractSessionID(ctx)

	assert.Equal(t, sessionID, extractedSessionID, "Expected SessionID to be %s, got %s", sessionID, extractedSessionID)
}

func TestExtractSessionID_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	extractedSessionID := ctxutil.ExtractSessionID(ctx)
	assert.Equal(t, "", extractedSessionID, "Expected SessionID to be \"\", got %s", extractedSessionID)
}

func TestContextChaining(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = ctxutil.InjectTxID(ctx, 67890)
	ctx = ctxutil.InjectSessionID(ctx, "xyz789")

	extractedTxID := ctxutil.ExtractTxID(ctx)
	extractedSessionID := ctxutil.ExtractSessionID(ctx)

	require.Equal(t, int64(67890), extractedTxID, "Expected TxID to be 67890, got %d", extractedTxID)
	require.Equal(t, "xyz789", extractedSessionID, "Expected SessionID to be \"xyz789\", got %s", extractedSessionID)
}

func TestInjectAndExtractTTL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctx = ctxutil.InjectTTL(ctx, "1s")

	extractedTxID := ctxutil.ExtractTTL(ctx)
	assert.Equal(t, "1s", extractedTxID, "Expected TxID to be %s, got %s", "1s", extractedTxID)
}
