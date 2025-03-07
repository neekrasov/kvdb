package ctxutil

import "context"

type TxID string

func InjectTxID(parent context.Context, value int64) context.Context {
	return context.WithValue(parent, TxID("tx"), value)
}

func ExtractTxID(ctx context.Context) int64 {
	val, ok := ctx.Value(TxID("tx")).(int64)
	if !ok {
		return 0
	}

	return val
}

type SessionID string

func InjectSessionID(parent context.Context, value string) context.Context {
	return context.WithValue(parent, SessionID("session"), value)
}

func ExtractSessionID(ctx context.Context) string {
	val, ok := ctx.Value(SessionID("session")).(string)
	if !ok {
		return ""
	}

	return val
}
