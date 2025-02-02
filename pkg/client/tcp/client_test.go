package tcp_test

import (
	"net"
	"syscall"
	"testing"
	"time"

	"github.com/neekrasov/kvdb/pkg/client/tcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockServer(t *testing.T, serverAdress, serverResponse string) {
	listener, err := net.Listen("tcp", serverAdress)
	require.NoError(t, err)

	go func() {
		for {
			connection, err := listener.Accept()
			if err != nil {
				return
			}

			_, err = connection.Read(make([]byte, 2048))
			require.NoError(t, err)

			_, err = connection.Write([]byte(serverResponse))
			require.NoError(t, err)
		}
	}()
}

func TestClient(t *testing.T) {
	t.Parallel()

	const (
		serverResponse = "hello client"
		serverAddress  = "localhost:2121"
	)

	mockServer(t, serverAddress, serverResponse)

	tests := map[string]struct {
		request          string
		client           func() *tcp.Client
		expectedResponse string
		expectedErr      error
	}{
		"client with incorrect server address": {
			request: "hello server",
			client: func() *tcp.Client {
				client, err := tcp.NewClient("localhost:2122")
				require.ErrorIs(t, err, syscall.ECONNREFUSED)
				return client
			},
			expectedResponse: serverResponse,
		},
		"client with small max message size": {
			request: "hello server",
			client: func() *tcp.Client {
				client, err := tcp.NewClient(serverAddress, tcp.WithClientBufferSize(1), tcp.WithClientIdleTimeout(time.Minute))
				require.NoError(t, err)
				return client
			},
			expectedErr: tcp.ErrSmallBufferSize,
		},
		"client without options": {
			request: "hello server",
			client: func() *tcp.Client {
				client, err := tcp.NewClient(serverAddress)
				require.NoError(t, err)
				return client
			},
			expectedResponse: serverResponse,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client := test.client()
			if client == nil {
				return
			}

			response, err := client.Send([]byte(test.request))

			if test.expectedErr != nil {
				assert.ErrorIs(t, err, test.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			if test.expectedResponse != "" {
				assert.Equal(t, test.expectedResponse, string(response[:len(test.expectedResponse)]))
			}

			client.Close()
		})
	}
}
