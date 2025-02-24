package replication

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/neekrasov/kvdb/internal/database/compute"
	"github.com/neekrasov/kvdb/internal/database/storage/wal"
	replMocks "github.com/neekrasov/kvdb/internal/mocks/replication"
	walMocks "github.com/neekrasov/kvdb/internal/mocks/wal"
	"github.com/neekrasov/kvdb/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewSlave(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	tests := []struct {
		name         string
		prepareMocks func(mockSegmentStorage *walMocks.SegmentStorage)
		expectError  bool
	}{
		{
			name: "Success - Create new slave",
			prepareMocks: func(mockSegmentStorage *walMocks.SegmentStorage) {
				mockSegmentStorage.On("List").Return([]int{1, 2, 3}, nil).Once()
			},
			expectError: false,
		},
		{
			name: "Error - List segments failed",
			prepareMocks: func(mockSegmentStorage *walMocks.SegmentStorage) {
				mockSegmentStorage.On("List").Return([]int{}, errors.New("list error")).Once()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNetClient := replMocks.NewNetClient(t)
			mockWAL := replMocks.NewWAL(t)
			mockSegmentStorage := walMocks.NewSegmentStorage(t)
			tt.prepareMocks(mockSegmentStorage)

			slave, err := NewSlave(mockNetClient, mockSegmentStorage, mockWAL, time.Second, 2)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, slave)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, slave)
			}

			mockSegmentStorage.AssertExpectations(t)
		})
	}
}

func TestSlave_Sync(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	tests := []struct {
		name         string
		prepareMocks func(mockNetClient *replMocks.NetClient, mockWAL *replMocks.WAL)
		expectError  bool
	}{
		{
			name: "Success - Sync with master empty result",
			prepareMocks: func(mockNetClient *replMocks.NetClient, mockWAL *replMocks.WAL) {
				masterRep := NewMasterResponse(false, 1, []byte{})
				buf := bytes.NewBuffer([]byte{})
				err := masterRep.Encode(buf)
				require.NoError(t, err)

				mockNetClient.On("Send", mock.Anything, mock.Anything).
					Return(buf.Bytes(), nil).Once()
			},
			expectError: false,
		},
		{
			name: "Success - Sync with master nil data",
			prepareMocks: func(mockNetClient *replMocks.NetClient, mockWAL *replMocks.WAL) {
				masterRep := NewMasterResponse(true, 1, []byte{})
				buf := bytes.NewBuffer([]byte{})
				err := masterRep.Encode(buf)
				require.NoError(t, err)

				mockNetClient.On("Send", mock.Anything, mock.Anything).
					Return(buf.Bytes(), nil).Once()
			},
			expectError: false,
		},
		{
			name: "Success - Sync with master nil data",
			prepareMocks: func(mockNetClient *replMocks.NetClient, mockWAL *replMocks.WAL) {
				masterRep := NewMasterResponse(true, 1, []byte{})
				buf := bytes.NewBuffer([]byte{})
				err := masterRep.Encode(buf)
				require.NoError(t, err)

				mockNetClient.On("Send", mock.Anything, mock.Anything).
					Return(buf.Bytes(), nil).Once()
			},
			expectError: false,
		},
		{
			name: "Success - Sync with master apply data",
			prepareMocks: func(mockNetClient *replMocks.NetClient, mockWAL *replMocks.WAL) {
				entry := wal.LogEntry{
					Operation: compute.CommandID(1),
					Args:      []string{"arg1", "arg2"},
				}

				var buf bytes.Buffer
				err := entry.Encode(&buf)
				require.NoError(t, err)

				masterRep := NewMasterResponse(true, 1, buf.Bytes())
				var buf2 bytes.Buffer
				err = masterRep.Encode(&buf2)
				require.NoError(t, err)

				mockNetClient.On("Send", mock.Anything, mock.Anything).
					Return(buf2.Bytes(), nil).Once()
				mockWAL.On("Flush", mock.Anything).Return(nil).Once()
			},
			expectError: false,
		},
		{
			name: "Error - Sync with master closed pipe",
			prepareMocks: func(mockNetClient *replMocks.NetClient, mockWAL *replMocks.WAL) {
				mockNetClient.On("Сonnect").Return(errors.New("test"))
				mockNetClient.On("Send", mock.Anything, mock.Anything).
					Return(nil, io.ErrClosedPipe).Once()
				mockNetClient.On("Send", mock.Anything, mock.Anything).
					Return(nil, errors.New("test")).Once()
			},
			expectError: false,
		},

		{
			name: "Error - invalid master response",
			prepareMocks: func(mockNetClient *replMocks.NetClient, mockWAL *replMocks.WAL) {
				masterRep := NewMasterResponse(true, 1, []byte("invalid"))
				buf := bytes.NewBuffer([]byte{})
				err := masterRep.Encode(buf)
				require.NoError(t, err)

				mockNetClient.On("Send", mock.Anything, mock.Anything).
					Return([]byte("invalid"), nil).Once()
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNetClient := replMocks.NewNetClient(t)
			mockWAL := replMocks.NewWAL(t)
			mockSegmentStorage := walMocks.NewSegmentStorage(t)
			mockSegmentStorage.On("List").Return([]int{1, 2, 3}, nil).Once()

			slave, err := NewSlave(mockNetClient, mockSegmentStorage, mockWAL, time.Millisecond, 2)
			require.NoError(t, err)
			require.NotNil(t, slave)

			tt.prepareMocks(mockNetClient, mockWAL)
			stream := slave.Stream()
			go func() {
				<-stream
			}()

			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				slave.sync(context.Background())
				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}()

			wg.Wait()

			mockNetClient.AssertExpectations(t)
			mockWAL.AssertExpectations(t)
		})
	}
}

func TestSlave_IsMaster(t *testing.T) {
	slave := &Slave{}
	assert.False(t, slave.IsMaster())
}
