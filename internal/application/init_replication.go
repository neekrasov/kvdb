package application

import (
	"errors"
	"fmt"
	"time"

	"github.com/neekrasov/kvdb/internal/config"
	"github.com/neekrasov/kvdb/internal/database/storage"
	"github.com/neekrasov/kvdb/internal/database/storage/replication"
	"github.com/neekrasov/kvdb/internal/database/storage/wal"
	"github.com/neekrasov/kvdb/internal/database/storage/wal/compressor"
	"github.com/neekrasov/kvdb/internal/database/storage/wal/filesystem"
	"github.com/neekrasov/kvdb/internal/database/storage/wal/segment"
	"github.com/neekrasov/kvdb/internal/delivery/tcp"
	sizeparser "github.com/neekrasov/kvdb/pkg/size_parser"
)

const (
	masterType = "master"
	slaveType  = "slave"
)

const (
	defaultReplicationSyncInterval = time.Second
	defaultMaxReplicasNumber       = 5
	defaultMaxSegmentSize          = 4 << 20
)

func initReplica(
	walInstance *wal.WAL,
	walCfg *config.WALConfig,
	replicationCfg *config.ReplicationConfig,
) (storage.Replica, error) {
	if replicationCfg == nil {
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
		size, err := sizeparser.ParseSize(walCfg.MaxSegmentSize)
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

	idleTimeout := syncInterval * 3
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

		var compressorInstance wal.Compressor
		if walCfg.Compression == "gzip" {
			compressorInstance = new(compressor.GzipCompressor)
		}

		iterator := wal.NewSegmentIterator(segmentStorage, compressorInstance)
		return replication.NewMaster(server, iterator), nil
	}

	var options []tcp.ClientOption
	options = append(options, tcp.WithClientIdleTimeout(idleTimeout))
	options = append(options, tcp.WithClientBufferSize(uint(maxMessageSize)))
	client, err := tcp.NewClient(masterAddress, options...)
	if err != nil {
		return nil, err
	}

	return replication.NewSlave(client, segmentStorage, walInstance, syncInterval, 10)
}
