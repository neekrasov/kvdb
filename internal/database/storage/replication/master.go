package replication

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/neekrasov/kvdb/pkg/logger"
	"go.uber.org/zap"
)

// Handler - type alias for the replication handler function, which processes requests from slaves.
type Handler = func(ctx context.Context, sessionID string, request []byte) []byte

type (
	// NetServer - interface for network server operations.
	NetServer interface {
		Close() error
		Start(ctx context.Context, handler Handler)
	}

	// Iterator - interface for iterating over data segments.
	Iterator interface {
		Next(int) ([]byte, error)
	}
)

// Master - struct representing the master server.
type Master struct {
	server   NetServer
	iterator Iterator
}

// NewMaster - constructor function that creates a new Master instance.
func NewMaster(
	server NetServer,
	iterator Iterator,
) *Master {
	return &Master{
		server:   server,
		iterator: iterator,
	}
}

// Start - starts the master server, handles replication requests from slaves, and sends responses.
func (m *Master) Start(ctx context.Context) {
	logger.Debug("replication master server was started")
	m.server.Start(ctx, func(ctx context.Context, _ string, requestData []byte) []byte {
		if ctx.Err() != nil {
			return nil
		}

		var request SlaveRequest
		err := request.Decode(bytes.NewReader(requestData))
		if err != nil && !errors.Is(err, io.EOF) {
			logger.Error("unable to decode replication request", zap.Error(err))
			return nil
		}

		logger.Debug("new slave request",
			zap.Int("segment_num", request.SegmentNum),
		)

		data, err := m.iterator.Next(request.SegmentNum)
		if err != nil {
			logger.Debug(
				"error while getting the next segment",
				zap.Error(err),
				zap.Int("segment_num", request.SegmentNum),
			)
		}

		response := MasterResponse{
			Succeed: err == nil,
			Data:    data,
		}

		var buffer bytes.Buffer
		if err := response.Encode(&buffer); err != nil {
			logger.Error("unable to encode replication response", zap.Error(err))
		}

		return buffer.Bytes()
	})

	if err := m.server.Close(); err != nil {
		logger.Warn("failed to close master http server", zap.Error(err))
	}
}

// IsMaster - returns true, indicating this instance is a master server.
func (m *Master) IsMaster() bool {
	return true
}
