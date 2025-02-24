package replication_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/neekrasov/kvdb/internal/database/storage/replication"
	mocks "github.com/neekrasov/kvdb/internal/mocks/replication"
	"github.com/neekrasov/kvdb/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMaster_Start(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	tests := []struct {
		name         string
		prepareMocks func(mockNetServer *mocks.NetServer, mockIterator *mocks.Iterator)
		expectError  bool
	}{
		{
			name: "Success - Start master server and handle request",
			prepareMocks: func(mockNetServer *mocks.NetServer, mockIterator *mocks.Iterator) {
				mockIterator.On("Next", 1).Return([]byte("segment data"), nil).Once()
				mockNetServer.On("Close").Return(nil).Once()

				mockNetServer.On("Start", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					ctx := args.Get(0).(context.Context)
					handler := args.Get(1).(replication.Handler)

					request := replication.SlaveRequest{SegmentNum: 1}
					var buffer bytes.Buffer
					err := request.Encode(&buffer)
					require.NoError(t, err)

					response := handler(ctx, buffer.Bytes())
					require.NotNil(t, response)

					var masterResponse replication.MasterResponse
					err = masterResponse.Decode(bytes.NewReader(response))
					require.NoError(t, err)
					assert.True(t, masterResponse.Succeed)
					assert.Equal(t, []byte("segment data"), masterResponse.Data)
				}).Once()
			},
			expectError: false,
		},
		{
			name: "Error - Iterator Next failed",
			prepareMocks: func(mockNetServer *mocks.NetServer, mockIterator *mocks.Iterator) {
				mockIterator.On("Next", 1).Return([]byte(nil), errors.New("iterator error")).Once()
				mockNetServer.On("Close").Return(nil).Once()
				mockNetServer.On("Start", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					ctx := args.Get(0).(context.Context)
					handler := args.Get(1).(replication.Handler)

					request := replication.SlaveRequest{SegmentNum: 1}
					var buffer bytes.Buffer
					err := request.Encode(&buffer)
					require.NoError(t, err)

					response := handler(ctx, buffer.Bytes())
					require.NotNil(t, response)

					var masterResponse replication.MasterResponse
					err = masterResponse.Decode(bytes.NewReader(response))
					require.NoError(t, err)
					assert.False(t, masterResponse.Succeed)
					assert.Nil(t, masterResponse.Data)
				}).Once()
			},
			expectError: false,
		},
		{
			name: "Error - Context canceled",
			prepareMocks: func(mockNetServer *mocks.NetServer, mockIterator *mocks.Iterator) {
				mockNetServer.On("Start", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					ctx, cancel := context.WithCancel(args.Get(0).(context.Context))
					cancel()

					handler := args.Get(1).(replication.Handler)

					request := replication.SlaveRequest{SegmentNum: 1}
					var buffer bytes.Buffer
					err := request.Encode(&buffer)
					require.NoError(t, err)

					response := handler(ctx, buffer.Bytes())
					assert.Nil(t, response)
				}).Once()

				mockNetServer.On("Close").Return(nil).Once()
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNetServer := mocks.NewNetServer(t)
			mockIterator := mocks.NewIterator(t)
			tt.prepareMocks(mockNetServer, mockIterator)

			master := replication.NewMaster(mockNetServer, mockIterator)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			master.Start(ctx)

			mockNetServer.AssertExpectations(t)
			mockIterator.AssertExpectations(t)
		})
	}
}

func TestMaster_IsMaster(t *testing.T) {
	master := &replication.Master{}
	assert.True(t, master.IsMaster())
}
