package tx

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
