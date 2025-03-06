package client_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/neekrasov/kvdb/internal/database"
	"github.com/neekrasov/kvdb/internal/database/compute"
	mocks "github.com/neekrasov/kvdb/internal/mocks/client"
	"github.com/neekrasov/kvdb/pkg/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	errPrefix = "[error]"
	okPrefix  = "[ok]"
)

func TestNewNetClient_Success(t *testing.T) {
	cfg := &client.Config{
		Address:              "localhost:8080",
		Username:             "user",
		Password:             "pass",
		MaxReconnectAttempts: 3,
	}

	ctx := context.Background()
	mockClientFactory := mocks.NewNetClientFactory(t)
	mockClient := mocks.NewNetClient(t)

	mockClientFactory.On("Make", cfg.Address, mock.Anything).
		Return(mockClient, nil)
	mockClient.On("Send", mock.Anything,
		[]byte(compute.CommandAUTH.Make(cfg.Username, cfg.Password))).
		Return([]byte(okPrefix), nil)

	kvdbClient, err := client.New(ctx, cfg, mockClientFactory)
	require.NoError(t, err, "NewNetClient should not return an error")
	assert.NotNil(t, kvdbClient, "Client should not be nil")

	mockClientFactory.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

func TestNewNetClient_ConnectionError(t *testing.T) {
	cfg := &client.Config{
		Address:              "localhost:8080",
		Username:             "user",
		Password:             "pass",
		MaxReconnectAttempts: 3,
	}

	ctx := context.Background()
	mockClientFactory := mocks.NewNetClientFactory(t)

	mockClientFactory.On("Make", cfg.Address, mock.Anything).
		Return(nil, errors.New("connection failed"))

	_, err := client.New(ctx, cfg, mockClientFactory)
	require.Error(t, err, "NewNetClient should return an error")
	assert.Contains(t, err.Error(), "initial connection failed", "Error message should contain 'initial connection failed'")

	mockClientFactory.AssertExpectations(t)
}

func TestRawWithRetries_Success(t *testing.T) {
	cfg := &client.Config{
		Address:              "localhost:8080",
		Username:             "user",
		Password:             "pass",
		MaxReconnectAttempts: 3,
	}

	ctx := context.Background()
	mockClientFactory := mocks.NewNetClientFactory(t)
	mockClient := mocks.NewNetClient(t)

	mockClientFactory.On("Make", cfg.Address, mock.Anything).Return(mockClient, nil)
	mockClient.On("Send", mock.Anything,
		[]byte(compute.CommandAUTH.Make(cfg.Username, cfg.Password))).
		Return([]byte(okPrefix), nil)

	kvdbClient, err := client.New(ctx, cfg, mockClientFactory)
	require.NoError(t, err, "NewNetClient should not return an error")

	getCmd := compute.CommandGET.Make("key")
	mockClient.On("Send", mock.Anything,
		[]byte(getCmd)).
		Return([]byte(database.WrapOK("value")), nil)

	res, err := kvdbClient.Raw(ctx, getCmd)
	require.NoError(t, err, "Raw should not return an error")
	assert.Equal(t, "value", res, "Response should be 'value'")

	mockClientFactory.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

func TestRawWithRetries_MaxReconnects(t *testing.T) {
	cfg := &client.Config{
		Address:              "localhost:8080",
		Username:             "user",
		Password:             "pass",
		MaxReconnectAttempts: 1,
		ReconnectBaseDelay:   1 * time.Microsecond,
	}

	ctx := context.Background()
	mockClientFactory := mocks.NewNetClientFactory(t)
	mockClient := mocks.NewNetClient(t)

	authCmd := compute.CommandAUTH.Make(cfg.Username, cfg.Password)
	mockClient.On("Close").Return(nil).Once()
	mockClientFactory.On("Make", cfg.Address, mock.Anything).Return(mockClient, nil).Once()
	mockClient.On("Send", mock.Anything, []byte(authCmd)).
		Return([]byte(okPrefix), nil).Once()

	mockClient.On("Send", mock.Anything,
		[]byte("GET key")).Return(nil, errors.New("connection failed")).Once()
	mockClientFactory.On("Make", cfg.Address, mock.Anything).Return(mockClient, nil).Once()
	mockClient.On("Send", mock.Anything,
		[]byte(authCmd)).
		Return([]byte(okPrefix), nil).Once()

	kvdbClient, err := client.New(ctx, cfg, mockClientFactory)
	require.NoError(t, err, "NewNetClient should not return an error")

	_, err = kvdbClient.Raw(ctx, "GET key")
	require.ErrorIs(t, err, client.ErrMaxReconnects)

	mockClientFactory.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

func TestAuthenticationFailure(t *testing.T) {
	cfg := &client.Config{
		Address:  "localhost:8080",
		Username: "user",
		Password: "pass",
	}

	ctx := context.Background()
	mockClientFactory := mocks.NewNetClientFactory(t)
	mockClient := mocks.NewNetClient(t)

	mockClientFactory.On("Make", cfg.Address, mock.Anything).Return(mockClient, nil)

	mockClient.On("Send", mock.Anything,
		[]byte(compute.CommandAUTH.Make(cfg.Username, cfg.Password))).
		Return([]byte(database.ErrAuthenticationRequired.Error()), nil)

	_, err := client.New(ctx, cfg, mockClientFactory)
	require.ErrorIs(t, err, client.ErrAuthenticationRequired)

	mockClientFactory.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

func TestClose(t *testing.T) {
	cfg := &client.Config{
		Address:              "localhost:8080",
		Username:             "user",
		Password:             "pass",
		MaxReconnectAttempts: 3,
	}

	ctx := context.Background()
	mockClientFactory := mocks.NewNetClientFactory(t)
	mockClient := mocks.NewNetClient(t)
	mockClientFactory.On("Make", cfg.Address, mock.Anything).Return(mockClient, nil)

	mockClient.On("Send", mock.Anything,
		[]byte(compute.CommandAUTH.Make(cfg.Username, cfg.Password))).
		Return([]byte(okPrefix), nil)

	kvdbClient, err := client.New(ctx, cfg, mockClientFactory)
	require.NoError(t, err, "NewNetClient should not return an error")

	mockClient.On("Close").Return(nil)
	err = kvdbClient.Close()
	require.NoError(t, err, "Close should not return an error")

	mockClientFactory.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}
