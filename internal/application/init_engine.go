package application

import (
	"errors"

	"github.com/neekrasov/kvdb/internal/config"
	"github.com/neekrasov/kvdb/internal/database/storage/engine"
)

func initEngine(cfg *config.EngineConfig) (*engine.Engine, error) {
	if cfg == nil {
		return nil, errors.New("empty engine config")
	}

	return engine.New(engine.WithPartitionNum(cfg.PartitionNum)), nil
}
