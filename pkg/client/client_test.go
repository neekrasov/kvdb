package client_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/neekrasov/kvdb/internal/database/models"
	mocks "github.com/neekrasov/kvdb/internal/mocks/client"
	"github.com/neekrasov/kvdb/pkg/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewClient_Success(t *testing.T) {
	cfg := &client.KVDBClientConfig{
		Address:              "localhost:8080",
		Username:             "user",
		Password:             "pass",
		MaxReconnectAttempts: 3,
	}

	mockClientFactory := mocks.NewClientFactory(t)
	mockClient := mocks.NewClient(t)
	mockClientFactory.On("Make", cfg.Address, mock.Anything).Return(mockClient, nil)

	mockClient.On("Send", mock.Anything, []byte(fmt.Sprintf("%s %s %s", models.CommandAUTH, cfg.Username, cfg.Password))).
		Return([]byte("OK"), nil)

	kvdbClient, err := client.New(cfg, mockClientFactory)
	require.NoError(t, err, "NewClient should not return an error")
	assert.NotNil(t, kvdbClient, "Client should not be nil")

	mockClientFactory.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

func TestNewClient_ConnectionError(t *testing.T) {
	cfg := &client.KVDBClientConfig{
		Address:              "localhost:8080",
		Username:             "user",
		Password:             "pass",
		MaxReconnectAttempts: 3,
	}

	mockClientFactory := mocks.NewClientFactory(t)
	mockClientFactory.On("Make", cfg.Address, mock.Anything).
		Return(nil, errors.New("connection failed"))

	_, err := client.New(cfg, mockClientFactory)
	require.Error(t, err, "NewClient should return an error")
	assert.Contains(t, err.Error(), "initial connection failed", "Error message should contain 'initial connection failed'")

	mockClientFactory.AssertExpectations(t)
}

func TestSendWithRetries_Success(t *testing.T) {
	cfg := &client.KVDBClientConfig{
		Address:              "localhost:8080",
		Username:             "user",
		Password:             "pass",
		MaxReconnectAttempts: 3,
	}

	mockClientFactory := mocks.NewClientFactory(t)
	mockClient := mocks.NewClient(t)

	mockClientFactory.On("Make", cfg.Address, mock.Anything).Return(mockClient, nil)
	mockClient.On("Send", mock.Anything, []byte(fmt.Sprintf("%s %s %s", models.CommandAUTH, cfg.Username, cfg.Password))).
		Return([]byte("OK"), nil)

	kvdbClient, err := client.New(cfg, mockClientFactory)
	require.NoError(t, err, "NewClient should not return an error")

	mockClient.On("Send", mock.Anything, []byte("GET key")).Return([]byte("value"), nil)

	res, err := kvdbClient.Send(context.Background(), "GET key")
	require.NoError(t, err, "Send should not return an error")
	assert.Equal(t, "value", res, "Response should be 'value'")

	mockClientFactory.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

func TestSendWithRetries_MaxReconnects(t *testing.T) {
	cfg := &client.KVDBClientConfig{
		Address:              "localhost:8080",
		Username:             "user",
		Password:             "pass",
		MaxReconnectAttempts: 1,
		ReconnectBaseDelay:   1 * time.Microsecond,
	}

	mockClientFactory := mocks.NewClientFactory(t)
	mockClient := mocks.NewClient(t)
	mockClient.On("Close").Return(nil).Once()
	mockClientFactory.On("Make", cfg.Address, mock.Anything).Return(mockClient, nil).Once()
	mockClient.On("Send", mock.Anything, []byte(fmt.Sprintf("%s %s %s", models.CommandAUTH, cfg.Username, cfg.Password))).
		Return([]byte("OK"), nil).Once()

	kvdbClient, err := client.New(cfg, mockClientFactory)
	require.NoError(t, err, "NewClient should not return an error")
	mockClient.On("Send", mock.Anything, []byte("GET key")).Return(nil, errors.New("connection failed")).Once()
	mockClientFactory.On("Make", cfg.Address, mock.Anything).Return(mockClient, nil).Once()
	mockClient.On("Send", mock.Anything, []byte(fmt.Sprintf("%s %s %s", models.CommandAUTH, cfg.Username, cfg.Password))).
		Return([]byte("OK"), nil).Once()

	_, err = kvdbClient.Send(context.Background(), "GET key")
	require.ErrorIs(t, err, client.ErrMaxReconnects)

	mockClientFactory.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

func TestAuthenticationFailure(t *testing.T) {
	cfg := &client.KVDBClientConfig{
		Address:  "localhost:8080",
		Username: "user",
		Password: "pass",
	}

	mockClientFactory := mocks.NewClientFactory(t)
	mockClient := mocks.NewClient(t)

	mockClientFactory.On("Make", cfg.Address, mock.Anything).Return(mockClient, nil)

	mockClient.On("Send", mock.Anything, []byte(fmt.Sprintf("%s %s %s", models.CommandAUTH, cfg.Username, cfg.Password))).
		Return([]byte(models.ErrAuthenticationRequired.Error()), nil)

	_, err := client.New(cfg, mockClientFactory)
	require.ErrorIs(t, err, client.ErrAuthenticationRequired)

	mockClientFactory.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

func TestClose(t *testing.T) {
	cfg := &client.KVDBClientConfig{
		Address:              "localhost:8080",
		Username:             "user",
		Password:             "pass",
		MaxReconnectAttempts: 3,
	}

	mockClientFactory := mocks.NewClientFactory(t)
	mockClient := mocks.NewClient(t)
	mockClientFactory.On("Make", cfg.Address, mock.Anything).Return(mockClient, nil)

	mockClient.On("Send", mock.Anything, []byte(fmt.Sprintf("%s %s %s", models.CommandAUTH, cfg.Username, cfg.Password))).
		Return([]byte("OK"), nil)

	kvdbClient, err := client.New(cfg, mockClientFactory)
	require.NoError(t, err, "NewClient should not return an error")

	mockClient.On("Close").Return(nil)
	err = kvdbClient.Close()
	require.NoError(t, err, "Close should not return an error")

	mockClientFactory.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}
