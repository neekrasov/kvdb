package application

import (
	"errors"

	"github.com/neekrasov/kvdb/internal/config"
	"github.com/neekrasov/kvdb/internal/database/storage/engine"
	"github.com/neekrasov/kvdb/pkg/logger"
	"go.uber.org/zap"
)

func initEngine(cfg *config.EngineConfig) (*engine.Engine, error) {
	if cfg == nil {
		return nil, errors.New("empty engine config")
	}

	var partitionNum int
	if cfg.PartitionNum == 0 {
		partitionNum = 1
	} else {
		partitionNum = cfg.PartitionNum
	}

	logger.Debug("init engine", zap.Int("partition_num", partitionNum))

	return engine.New(engine.WithPartitionNum(cfg.PartitionNum)), nil
}
