package replication

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"syscall"
	"time"

	"github.com/neekrasov/kvdb/internal/database/storage/wal"
	"github.com/neekrasov/kvdb/pkg/logger"
	"go.uber.org/zap"
)

const (
	defaultSyncRetryNum      = 10
	defaultSyncRetryDuration = time.Second
)

// Stream - channel type for sending log entries
type Stream chan []wal.LogEntry

// NetClient - interface for network client operations.
type NetClient interface {
	Send(ctx context.Context, request []byte) ([]byte, error)
	Сonnect() error
}

// WAL - interface for write-ahead log operations.
type WAL interface {
	Flush(batch []wal.WriteEntry) error
}

// Slave - struct representing the slave server.
type Slave struct {
	client            NetClient
	stream            Stream
	walInstance       WAL
	lastSegmentNum    int
	syncInterval      time.Duration
	syncRetryNum      int
	syncRetryDuration time.Duration
}

// NewSlave - constructor function that creates a new Slave instance.
func NewSlave(
	client NetClient,
	segmentStorage wal.SegmentStorage,
	walInstance WAL,
	syncInterval time.Duration,
	syncRetryNum int,
) (*Slave, error) {
	ids, err := segmentStorage.List()
	if err != nil {
		return nil, fmt.Errorf("list segments failed: %w", err)
	}

	var lastSegmentNum int
	if len(ids) > 0 {
		lastSegmentNum = ids[len(ids)-1]
	}

	slave := &Slave{
		client:            client,
		lastSegmentNum:    lastSegmentNum,
		syncInterval:      syncInterval,
		walInstance:       walInstance,
		syncRetryNum:      defaultSyncRetryNum,
		syncRetryDuration: defaultSyncRetryDuration,
		stream:            make(Stream),
	}

	if syncRetryNum != 0 {
		slave.syncRetryNum = syncRetryNum
	}

	return slave, nil
}

// Start - starts the slave's synchronization process, periodically syncing data with the master.
func (s *Slave) Start(ctx context.Context) {
	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	logger.Debug("sync started",
		zap.Stringer("time", time.Now()),
		zap.Stringer("interval", s.syncInterval),
	)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			logger.Debug("synchronize", zap.Stringer("time", time.Now()))

			if err := s.sync(ctx); err != nil {
				logger.Error("slave sync failed", zap.Error(err))
			}
		}
	}
}

// sync - handles the process of syncing the slave with the master server by sending a request and receiving data.
func (s *Slave) sync(ctx context.Context) error {
	var request bytes.Buffer
	err := NewSlaveRequest(s.lastSegmentNum).Encode(&request)
	if err != nil {
		return fmt.Errorf("failed to make slave request (s.num %d): %w", s.lastSegmentNum, err)
	}

	var resp []byte
	for i := range s.syncRetryNum {
		resp, err = s.client.Send(ctx, request.Bytes())
		if err != nil && !errors.Is(err, io.EOF) {
			if errors.Is(err, syscall.EPIPE) ||
				errors.Is(err, io.ErrClosedPipe) ||
				errors.Is(err, syscall.ECONNRESET) {
				logger.Warn("master unavailable, retry", zap.Error(err), zap.Int("retry_num", i))
				time.Sleep(s.syncRetryDuration)
				if err := s.client.Сonnect(); err != nil {
					logger.Warn("create connection to master failed", zap.Error(err))
				}
				continue
			}

			return fmt.Errorf("error sending slave request (s.num %d): %w", s.lastSegmentNum, err)
		}

		break
	}

	var response MasterResponse
	err = response.Decode(bytes.NewReader(resp))
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("decode master response failed (s.num %d): %w", s.lastSegmentNum, err)
	}
	logger.Debug("master response",
		zap.Bool("succeed", response.Succeed),
		zap.Int("segment_num", s.lastSegmentNum),
	)

	if response.Succeed {
		if err := s.applySegment(response.Data); err != nil {
			return fmt.Errorf("failed to upply segment data: %w", err)
		}
		logger.Debug("apply segment success",
			zap.Int("segment_num", s.lastSegmentNum))
		s.lastSegmentNum += 1
	}

	return nil
}

// applySegment - applies the received segment data to the slave's write-ahead log (WAL) and stream.
func (s *Slave) applySegment(payload []byte) error {
	if len(payload) == 0 {
		return nil
	}

	var (
		logEntries   []wal.LogEntry
		writeEntries []wal.WriteEntry
	)
	buffer := bytes.NewBuffer(payload)
	for buffer.Len() > 0 {
		var request wal.LogEntry
		if err := request.Decode(buffer); err != nil {
			return fmt.Errorf("unable to parse log entry: %w", err)
		}

		logEntries = append(logEntries, request)
		writeEntries = append(writeEntries, wal.NewWriteEntry(
			request.LSN, request.Operation, request.Args,
		))
	}

	if err := s.walInstance.Flush(writeEntries); err != nil {
		return fmt.Errorf("flush segment to wal failed: %w", err)
	}

	s.stream <- logEntries

	return nil
}

// Stream - returns the stream channel for the slave's log entries.
func (s *Slave) Stream() Stream {
	return s.stream
}

// IsMaster - returns false, indicating this instance is a slave server.
func (m *Slave) IsMaster() bool {
	return false
}
