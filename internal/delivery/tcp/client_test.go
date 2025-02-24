package tcp

import (
	"context"
	"errors"
	"net"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAddress = "localhost:12345"

func TestNewClient(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		ln, err := net.Listen("tcp", testAddress)
		require.NoError(t, err)
		defer ln.Close()

		go func() {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}()

		client, err := NewClient(testAddress)
		require.NoError(t, err)
		defer client.Close()

		assert.NotNil(t, client.connection)
	})

	t.Run("connection failed", func(t *testing.T) {
		client, err := NewClient("localhost:9999")
		require.Error(t, err)
		assert.Nil(t, client)
	})
}

func TestClient_Send(t *testing.T) {
	t.Run("successful send and receive", func(t *testing.T) {
		ln, err := net.Listen("tcp", testAddress)
		require.NoError(t, err)
		defer ln.Close()

		go func() {
			conn, err := ln.Accept()
			require.NoError(t, err)
			defer conn.Close()

			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			require.NoError(t, err)
			assert.Equal(t, "hello", string(buf[:n]))

			_, err = conn.Write([]byte("world"))
			require.NoError(t, err)
		}()

		client, err := NewClient(testAddress)
		require.NoError(t, err)
		defer client.Close()

		response, err := client.Send(context.Background(), []byte("hello"))
		require.NoError(t, err)
		assert.Equal(t, "world", string(response))
	})

	t.Run("timeout", func(t *testing.T) {
		ln, err := net.Listen("tcp", testAddress)
		require.NoError(t, err)
		defer ln.Close()

		go func() {
			conn, err := ln.Accept()
			require.NoError(t, err)
			defer conn.Close()

			time.Sleep(200 * time.Millisecond)
		}()

		client, err := NewClient(testAddress, WithClientIdleTimeout(100*time.Millisecond))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.Send(context.Background(), []byte("hello"))
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrTimeout))
	})

	t.Run("small buffer size", func(t *testing.T) {
		ln, err := net.Listen("tcp", testAddress)
		require.NoError(t, err)
		defer ln.Close()

		go func() {
			conn, err := ln.Accept()
			require.NoError(t, err)
			defer conn.Close()

			buf := make([]byte, 5)
			n, err := conn.Read(buf)
			require.NoError(t, err)
			assert.Equal(t, "hell", string(buf[:n]))

			_, err = conn.Write([]byte("hello world"))
			require.NoError(t, err)
		}()

		client, err := NewClient(testAddress, WithClientBufferSize(5))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.Send(context.Background(), []byte("hell"))
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrSmallBufferSize))

		_, err = client.Send(context.Background(), []byte("hello world"))
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrSmallBufferSize))
	})

	t.Run("connection closed", func(t *testing.T) {
		_, err := NewClient(testAddress)
		require.Error(t, err)
		assert.ErrorIs(t, err, syscall.ECONNREFUSED)
	})
}

func TestClient_Close(t *testing.T) {
	t.Run("successful close", func(t *testing.T) {
		ln, err := net.Listen("tcp", testAddress)
		require.NoError(t, err)
		defer ln.Close()

		go func() {
			conn, err := ln.Accept()
			require.NoError(t, err)
			conn.Close()
		}()

		client, err := NewClient(testAddress)
		require.NoError(t, err)

		err = client.Close()
		require.NoError(t, err)
		assert.Nil(t, client.connection)
	})

	t.Run("double close", func(t *testing.T) {
		ln, err := net.Listen("tcp", testAddress)
		require.NoError(t, err)
		defer ln.Close()

		go func() {
			conn, err := ln.Accept()
			require.NoError(t, err)
			conn.Close()
		}()

		client, err := NewClient(testAddress)
		require.NoError(t, err)

		err = client.Close()
		require.NoError(t, err)
		err = client.Close()
		require.NoError(t, err)
	})
}
