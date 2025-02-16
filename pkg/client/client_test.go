package client

import (
	"bytes"
	"errors"
	"io"
	"syscall"
	"testing"
	"time"

	"github.com/chzyer/readline"
	mocks "github.com/neekrasov/kvdb/internal/mocks/client"
	"github.com/neekrasov/kvdb/pkg/client/tcp"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cfg      *KVDBClientConfig
		errorMsg string
	}{
		{
			name: "Valid configuration",
			cfg: &KVDBClientConfig{
				Address:        "localhost:6379",
				IdleTimeout:    10 * time.Second,
				MaxMessageSize: "1MB",
				Username:       "root",
				Password:       "root",
			},
			errorMsg: "init tcp client failed: failed to dial: dial tcp [::1]:6379: connect: connection refused",
		},
		{
			name:     "Empty address",
			cfg:      &KVDBClientConfig{},
			errorMsg: "empty address",
		},
		{
			name: "Invalid max message size",
			cfg: &KVDBClientConfig{
				Address:        "localhost:6379",
				IdleTimeout:    10 * time.Second,
				Username:       "root",
				Password:       "root",
				MaxMessageSize: "invalid_size",
			},
			errorMsg: "parse max message size 'invalid_size' failed",
		},
		{
			name: "Invalid username or password",
			cfg: &KVDBClientConfig{
				Address:        "localhost:6379",
				IdleTimeout:    10 * time.Second,
				Username:       "root",
				MaxMessageSize: "invalid_size",
			},
			errorMsg: "username and password must be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewClient(tt.cfg)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

func TestSend(t *testing.T) {
	mockTCPClient := new(mocks.Client)
	clientInstance := &KVDBClient{client: mockTCPClient}

	mockTCPClient.On("Send", []byte("GET key")).Return([]byte("value"), nil)

	response, err := clientInstance.Send("GET key")
	assert.NoError(t, err)
	assert.Equal(t, "value", response)
}

func TestSend_Error(t *testing.T) {
	mockTCPClient := new(mocks.Client)
	clientInstance := &KVDBClient{client: mockTCPClient}

	sendErr := errors.New("connection error")
	mockTCPClient.On("Send", []byte("GET key")).Return([]byte{}, sendErr)

	response, err := clientInstance.Send("GET key")
	assert.Error(t, err)
	assert.Equal(t, "", response)
	assert.Equal(t, sendErr, err)
}

func TestClose(t *testing.T) {
	mockTCPClient := new(mocks.Client)
	clientInstance := &KVDBClient{client: mockTCPClient}

	mockTCPClient.On("Close").Return(nil)

	err := clientInstance.Close()
	assert.NoError(t, err)
}

func TestCloseNilClient(t *testing.T) {
	clientInstance := &KVDBClient{client: nil}
	err := clientInstance.Close()
	assert.NoError(t, err)
}

func TestCLI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        string
		mockSendResp []byte
		mockSendErr  error
		expectedOut  string
		expectedErr  error
		writer       MockWriter
	}{
		{
			name:         "normal input",
			input:        "ping\nexit\n",
			mockSendResp: []byte("pong"),
			mockSendErr:  nil,
			expectedOut:  "pong",
			expectedErr:  nil,
		},
		{
			name:         "readline interrupt",
			input:        "\x03",
			mockSendResp: nil,
			mockSendErr:  nil,
			expectedOut:  "",
			expectedErr:  nil,
		},
		{
			name:         "syscall.EPIPE error",
			input:        "ping\nexit\n",
			mockSendResp: nil,
			mockSendErr:  syscall.EPIPE,
			expectedOut:  "",
			expectedErr:  syscall.EPIPE,
		},
		{
			name:         "tcp.ErrTimeout error",
			input:        "ping\nexit\n",
			mockSendResp: nil,
			mockSendErr:  tcp.ErrTimeout,
			expectedOut:  "",
			expectedErr:  tcp.ErrTimeout,
		},
		{
			name:         "syscall.ECONNRESET error",
			input:        "ping\nexit\n",
			mockSendResp: nil,
			mockSendErr:  syscall.ECONNRESET,
			expectedOut:  "",
			expectedErr:  syscall.ECONNRESET,
		},
		{
			name:         "generic send error",
			input:        "ping\nexit\n",
			mockSendResp: nil,
			mockSendErr:  errors.New("some network error"),
			expectedOut:  "error: sending query failed: some network error\n",
			expectedErr:  nil,
		},
		{
			name:         "write error when reading stdin",
			input:        "ping\nexit\n",
			mockSendResp: nil,
			mockSendErr:  nil,
			writer:       new(mockWriter),
			expectedErr:  ErrWriteLineFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockTCPClient := new(mocks.Client)
			clientInstance := &KVDBClient{client: mockTCPClient}

			mockTCPClient.On("Send", []byte("ping")).Return(tt.mockSendResp, tt.mockSendErr)
			mockTCPClient.On("Close").Return(nil)

			input := bytes.NewBufferString(tt.input)
			var output MockWriter
			if tt.writer != nil {
				output = tt.writer
				_ = output.String()
			} else {
				output = new(bytes.Buffer)
			}

			rl, err := readline.NewEx(&readline.Config{
				Prompt:      "$ ",
				Stdin:       io.NopCloser(input),
				Stdout:      output,
				HistoryFile: "",
			})
			assert.NoError(t, err)
			defer rl.Close()

			err = clientInstance.CLI(rl)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectedOut != "" {
				assert.Contains(t, output.String(), tt.expectedOut)
			}
		})
	}
}
