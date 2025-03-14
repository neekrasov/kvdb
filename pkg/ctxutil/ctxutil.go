package ctxutil

import "context"

type ctxKey int

const (
	sessionIDKey ctxKey = iota
	txIDKey
	ttlKey
)

// InjectTxID - adds a Transaction ID (txID) to the context.
func InjectTxID(parent context.Context, value int64) context.Context {
	return context.WithValue(parent, txIDKey, value)
}

// ExtractTxID - retrieves the Transaction ID (txID) from the context. Returns 0 if not found or invalid.
func ExtractTxID(ctx context.Context) int64 {
	val, ok := ctx.Value(txIDKey).(int64)
	if !ok {
		return 0
	}

	return val
}

// InjectSessionID - adds a Session ID to the context.
func InjectSessionID(parent context.Context, value string) context.Context {
	return context.WithValue(parent, sessionIDKey, value)
}

// ExtractSessionID - retrieves the Session ID from the context.
// Returns empty string if not found or invalid.
func ExtractSessionID(ctx context.Context) string {
	val, ok := ctx.Value(sessionIDKey).(string)
	if !ok {
		return ""
	}

	return val
}

// InjectTTL - adds a TTL (Time-To-Live) value to the context.
func InjectTTL(ctx context.Context, ttl string) context.Context {
	return context.WithValue(ctx, ttlKey, ttl)
}

// ExtractTTL - retrieves the TTL value from the context.
// Returns empty string if not found.
func ExtractTTL(ctx context.Context) string {
	ttl, _ := ctx.Value(ttlKey).(string)
	return ttl
}
