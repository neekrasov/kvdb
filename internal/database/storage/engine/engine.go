package engine

import (
	"context"
	"hash/fnv"

	"github.com/neekrasov/kvdb/pkg/ctxutil"
	"github.com/neekrasov/kvdb/pkg/logger"
	pkgsync "github.com/neekrasov/kvdb/pkg/sync"
	"go.uber.org/zap"
)

// Engine - abstract data storage engine.
type Engine struct {
	partitions []*partitionMap
}

// New - creates a new instance of Engine.
func New(options ...Option) *Engine {
	e := new(Engine)

	for _, option := range options {
		option(e)
	}

	if len(e.partitions) == 0 {
		e.partitions = make([]*partitionMap, 1)
		e.partitions[0] = newPartMap()
	}

	return e
}

// Set - set stores a key-value pair in memory.
func (e *Engine) Set(ctx context.Context, key, value string) {
	txID := ctxutil.ExtractTxID(ctx)
	sessionID := ctxutil.ExtractSessionID(ctx)

	n, part := e.part(txID, sessionID, key)
	part.Set(key, value)

	logger.Debug(
		"successfull set query",
		zap.Int64("tx", txID),
		zap.String("session", sessionID),
		zap.Int("part", n),
	)
}

// Get - retrieves the value associated with a key.
func (e *Engine) Get(ctx context.Context, key string) (string, bool) {
	txID := ctxutil.ExtractTxID(ctx)
	sessionID := ctxutil.ExtractSessionID(ctx)

	n, part := e.part(txID, sessionID, key)
	val, found := part.Get(key)

	logger.Debug(
		"successfull get query",
		zap.Int64("tx", txID),
		zap.String("session", sessionID),
		zap.Int("part", n),
	)

	return val, found
}

// Watch - watches the key and returns the value if it has changed.
func (e *Engine) Watch(ctx context.Context, key string) pkgsync.FutureString {
	txID := ctxutil.ExtractTxID(ctx)
	sessionID := ctxutil.ExtractSessionID(ctx)

	n, part := e.part(txID, sessionID, key)
	logger.Debug(
		"successfull watch query",
		zap.Int64("tx", txID),
		zap.String("session", sessionID),
		zap.Int("part", n),
	)

	return part.Watch(ctx, key)
}

// Del - removes a key-value pair from memory.
func (e *Engine) Del(ctx context.Context, key string) error {
	txID := ctxutil.ExtractTxID(ctx)
	sessionID := ctxutil.ExtractSessionID(ctx)

	n, part := e.part(txID, sessionID, key)
	err := part.Del(key)
	if err != nil {
		logger.Debug("del query failed",
			zap.Int64("tx", txID),
			zap.String("session", sessionID),
			zap.Int("part", n),
			zap.Error(err),
		)
	} else {
		logger.Debug("successfull get query",
			zap.Int64("tx", txID),
			zap.String("session", sessionID),
			zap.Int("part", n),
		)
	}

	return err
}

// part - returns the partition for a given key based on hashing.
func (e *Engine) part(txID int64, sessionID string, key string) (int, *partitionMap) {
	hash := fnv.New32a()
	if _, err := hash.Write([]byte(key)); err != nil {
		logger.Error(
			"hash key failed",
			zap.String("key", key),
			zap.Int64("tx", txID),
			zap.String("session", sessionID),
			zap.Error(err),
		)
		return 0, nil
	}

	num := int(hash.Sum32()) % len(e.partitions)
	return num, e.partitions[num]
}
