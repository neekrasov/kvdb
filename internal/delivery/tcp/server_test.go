package tcp

import (
	"context"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/neekrasov/kvdb/internal/database"
	mocks "github.com/neekrasov/kvdb/internal/mocks/tcp"
	"github.com/neekrasov/kvdb/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	db := &database.Database{}
	server := NewServer(db, WithServerMaxConnectionsNumber(5), WithServerBufferSize(512), WithServerIdleTimeout(10*time.Second))

	assert.NotNil(t, server)
	assert.Equal(t, uint(512), server.bufferSize)
	assert.Equal(t, 10*time.Second, server.idleTimeout)
	assert.Equal(t, uint(5), server.maxConnections)
	assert.NotNil(t, server.semaphore)
}

func TestServer_StartWithInvalidAddress(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	db := &database.Database{}
	server := NewServer(db)

	err := server.Start(context.Background(), "")
	assert.Error(t, err)
	assert.Equal(t, "empty address", err.Error())
}

func TestServer(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	databaseMock := mocks.NewQueryHandler(t)
	databaseMock.EXPECT().HandleQuery("client-1").Return("hello-client-1")
	databaseMock.EXPECT().HandleQuery("client-2").Return("hello-client-2")
	server := NewServer(databaseMock, WithServerIdleTimeout(time.Minute))

	serverAddress := "localhost:22222"
	go func() {
		server.Start(ctx, serverAddress)
	}()

	time.Sleep(100 * time.Millisecond)

	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()

		conn, clientErr := net.Dial("tcp", serverAddress)
		require.NoError(t, clientErr)

		_, clientErr = conn.Write([]byte("client-1"))
		require.NoError(t, clientErr)

		buffer := make([]byte, 1024)
		size, clientErr := conn.Read(buffer)
		require.NoError(t, clientErr)

		clientErr = conn.Close()
		require.NoError(t, clientErr)

		assert.Equal(t, "hello-client-1", string(buffer[:size]))
	}()

	go func() {
		defer wg.Done()

		conn, clientErr := net.Dial("tcp", serverAddress)
		require.NoError(t, clientErr)

		_, clientErr = conn.Write([]byte("client-2"))
		require.NoError(t, clientErr)

		buffer := make([]byte, 1024)
		size, clientErr := conn.Read(buffer)
		require.NoError(t, clientErr)

		clientErr = conn.Close()
		require.NoError(t, clientErr)

		assert.Equal(t, "hello-client-2", string(buffer[:size]))
	}()

	go func() {
		defer wg.Done()

		conn, clientErr := net.Dial("tcp", serverAddress)
		require.NoError(t, clientErr)

		writeBuf := make([]byte, 4<<10)
		writeBuf[0] = 1
		writeBuf[4<<10-1] = 1

		_, clientErr = conn.Write(writeBuf)
		require.NoError(t, clientErr)

		buffer := make([]byte, 1024)
		_, clientErr = conn.Read(buffer)
		require.ErrorIs(t, clientErr, io.EOF)

		require.NoError(t, conn.Close())
	}()

	wg.Wait()
}
