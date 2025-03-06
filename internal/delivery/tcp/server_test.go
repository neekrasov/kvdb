package tcp

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/neekrasov/kvdb/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	server, err := NewServer("0.0.0.0:3232", WithServerMaxConnectionsNumber(5), WithServerBufferSize(512), WithServerIdleTimeout(10*time.Second))

	assert.NoError(t, err)
	assert.NotNil(t, server)
	assert.Equal(t, uint(512), server.bufferSize)
	assert.Equal(t, 10*time.Second, server.idleTimeout)
	assert.Equal(t, uint(5), server.maxConnections)
	assert.NotNil(t, server.semaphore)
}

func TestServer_StartWithInvalidAddress(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	_, err := NewServer("")
	assert.Error(t, err)
	assert.Equal(t, "empty address", err.Error())
}

func TestServer(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverAddress := "localhost:22223"
	server, err := NewServer(serverAddress, WithServerIdleTimeout(time.Minute))
	require.NoError(t, err)
	defer server.Close()

	go func() {
		server.Start(ctx, func(ctx context.Context, _ string, data []byte) []byte {
			return []byte("[ok] " + string(data))
		})
	}()

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()

		firstConn, clientErr := net.Dial("tcp", serverAddress)
		require.NoError(t, clientErr)

		_, clientErr = firstConn.Write([]byte("client-1"))
		require.NoError(t, clientErr)

		buffer := make([]byte, 1024)
		size, clientErr := firstConn.Read(buffer)
		require.NoError(t, clientErr)
		time.Sleep(time.Millisecond * 4)

		clientErr = firstConn.Close()
		require.NoError(t, clientErr)

		assert.Equal(t, "[ok] client-1", string(buffer[:size]))
	}()

	go func() {
		defer wg.Done()

		secondConn, clientErr := net.Dial("tcp", serverAddress)
		require.NoError(t, clientErr)

		_, clientErr = secondConn.Write([]byte("client-2"))
		require.NoError(t, clientErr)

		buffer := make([]byte, 1024)
		size, clientErr := secondConn.Read(buffer)
		require.NoError(t, clientErr)

		time.Sleep(time.Millisecond * 4)

		clientErr = secondConn.Close()
		require.NoError(t, clientErr)

		assert.Equal(t, "[ok] client-2", string(buffer[:size]))
	}()

	time.Sleep(time.Millisecond)
	require.Equal(t, int32(2), server.ActiveConnections())

	wg.Wait()
}
