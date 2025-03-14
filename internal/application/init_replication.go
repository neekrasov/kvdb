package application

import (
	"errors"
	"fmt"
	"time"

	"github.com/neekrasov/kvdb/internal/config"
	"github.com/neekrasov/kvdb/internal/database/compression"
	"github.com/neekrasov/kvdb/internal/database/storage"
	"github.com/neekrasov/kvdb/internal/database/storage/replication"
	"github.com/neekrasov/kvdb/internal/database/storage/wal"
	"github.com/neekrasov/kvdb/internal/database/storage/wal/filesystem"
	"github.com/neekrasov/kvdb/internal/database/storage/wal/segment"
	"github.com/neekrasov/kvdb/internal/delivery/tcp"
	"github.com/neekrasov/kvdb/pkg/logger"
	"github.com/neekrasov/kvdb/pkg/sizeutil"
	"go.uber.org/zap"
)

const (
	masterType = "master"
	slaveType  = "slave"
)

const (
	defaultReplicationSyncInterval = time.Second
	defaultMaxReplicasNumber       = 5
	defaultMaxSegmentSize          = 4 << 10 // 4KB
)

func initReplica(
	walInstance *wal.WAL,
	walCfg *config.WALConfig,
	replicationCfg *config.ReplicationConfig,
) (storage.Replica, error) {
	if replicationCfg == nil {
		logger.Warn("empty replication config")
		return nil, nil
	} else if walCfg == nil || walInstance == nil {
		return nil, errors.New("replication without wal")
	}

	rType := replicationCfg.ReplicaType
	if rType == "" {
		return nil, errors.New("empty replica type")
	} else if rType != masterType && rType != slaveType {
		return nil, fmt.Errorf("invalud replica type '%s'", rType)
	}

	masterAddress := replicationCfg.MasterAddress
	if masterAddress == "" {
		return nil, errors.New("empty master address")
	}

	syncInterval := defaultReplicationSyncInterval
	if replicationCfg.SyncInterval != 0 {
		syncInterval = replicationCfg.SyncInterval
	}

	maxMessageSize := defaultMaxSegmentSize
	if walCfg.MaxSegmentSize != "" {
		size, err := sizeutil.ParseSize(walCfg.MaxSegmentSize)
		if err != nil {
			return nil, fmt.Errorf("parse max segment size failed: %w", err)
		}
		maxMessageSize = size
	}

	segmentStorage, err := segment.NewFileSegmentStorage(
		new(filesystem.LocalFileSystem), walCfg.DataDir)
	if err != nil {
		return nil, err
	}

	idleTimeout := syncInterval * 3 // TODO: move to config?
	if replicationCfg.ReplicaType == masterType {
		maxReplicasNumber := defaultMaxReplicasNumber
		if replicationCfg.MaxReplicasNumber != 0 {
			maxReplicasNumber = replicationCfg.MaxReplicasNumber
		}

		server, err := tcp.NewServer(
			replicationCfg.MasterAddress,
			tcp.WithServerIdleTimeout(idleTimeout),
			tcp.WithServerBufferSize(uint(maxMessageSize)),
			tcp.WithServerMaxConnectionsNumber(uint(maxReplicasNumber)),
		)
		if err != nil {
			return nil, err
		}

		var compressor compression.Compressor
		if walCfg.Compression != "" {
			compressor, err = compression.New(walCfg.Compression)
			if err != nil {
				return nil, err
			}
		}

		logger.Debug("init master replica",
			zap.Int("max_replicas_number", maxReplicasNumber),
			zap.Int("max_segment_size", maxMessageSize),
			zap.String("master_address", masterAddress),
			zap.Stringer("idle_timeout", idleTimeout),
		)

		iterator := wal.NewSegmentIterator(segmentStorage, compressor)
		return replication.NewMaster(server, iterator), nil
	}

	var options []tcp.ClientOption
	options = append(options, tcp.WithClientIdleTimeout(idleTimeout))
	options = append(options, tcp.WithClientBufferSize(uint(maxMessageSize)))
	client, err := tcp.NewClient(masterAddress, options...)
	if err != nil {
		return nil, err
	}

	logger.Debug("init slave replica",
		zap.Int("max_segment_size", maxMessageSize),
		zap.Stringer("sync_interval", syncInterval),
		zap.Stringer("idle_timeout", idleTimeout),
	)

	return replication.NewSlave(client, segmentStorage, walInstance, syncInterval, 10)
}
