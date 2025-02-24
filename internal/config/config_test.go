package config_test

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/neekrasov/kvdb/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		content     string
		format      string
		expected    config.Config
		expectError bool
	}{
		{
			name:   "Valid YAML config",
			format: "yaml",
			content: `
engine:
  type: "in_memory"
network:
  address: "127.0.0.1:3221"
  max_connections: 200
  max_message_size: "5KB"
  idle_timeout: 6m
logging:
  level: "debug"
  output: "/log/output_test.log"
`,
			expected: config.Config{
				Engine: &config.EngineConfig{
					Type: "in_memory",
				},
				Network: &config.NetworkConfig{
					Address:        "127.0.0.1:3221",
					MaxConnections: 200,
					MaxMessageSize: "5KB",
					IdleTimeout:    6 * time.Minute,
				},
				Logging: &config.LoggingConfig{
					Level:  "debug",
					Output: "/log/output_test.log",
				},
			},
			expectError: false,
		},
		{
			name:   "Invalid YAML config (Invalid time format)",
			format: "yaml",
			content: `
engine:
  type: "in_memory"
network:
  address: "127.0.0.1:3221"
  max_connections: 200
  max_message_size: "5KB"
  idle_timeout: "invalid-time"
logging:
  level: "debug"
  output: "/log/output_test.log"
`,
			expected:    config.Config{},
			expectError: true,
		},
		{
			name:   "Valid JSON config",
			format: "json",
			content: `{
				"engine": {
					"type": "in_memory"
				},
				"network": {
					"address": "127.0.0.1:3221",
					"max_connections": 200,
					"max_message_size": "5KB",
					"idle_timeout": "360s"
				},
				"logging": {
					"level": "debug",
					"output": "/log/output_test.log"
				}
			}`,
			expected: config.Config{
				Engine: &config.EngineConfig{
					Type: "in_memory",
				},
				Network: &config.NetworkConfig{
					Address:        "127.0.0.1:3221",
					MaxConnections: 200,
					MaxMessageSize: "5KB",
					IdleTimeout:    6 * time.Minute,
				},
				Logging: &config.LoggingConfig{
					Level:  "debug",
					Output: "/log/output_test.log",
				},
			},
			expectError: false,
		},
		{
			name:   "Invalid JSON config (Invalid time format)",
			format: "json",
			content: `{
				"engine": {
					"type": "in_memory"
				},
				"network": {
					"address": "127.0.0.1:3221",
					"max_connections": 200,
					"max_message_size": "5KB",
					"idle_timeout": "invalid-time"
				},
				"logging": {
					"level": "debug",
					"output": "/log/output_test.log"
				}
			}`,
			expected:    config.Config{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockReader := bytes.NewReader([]byte(tt.content))
			cfg, err := config.ParseConfig(io.NopCloser(mockReader))
			if !tt.expectError {
				require.NoError(t, err)
				assert.Equal(t, tt.expected.Engine.Type, cfg.Engine.Type)
				assert.Equal(t, tt.expected.Network.Address, cfg.Network.Address)
				assert.Equal(t, tt.expected.Network.MaxConnections, cfg.Network.MaxConnections)
				assert.Equal(t, tt.expected.Network.MaxMessageSize, cfg.Network.MaxMessageSize)
				assert.Equal(t, tt.expected.Network.IdleTimeout, cfg.Network.IdleTimeout)
				assert.Equal(t, tt.expected.Logging.Level, cfg.Logging.Level)
				assert.Equal(t, tt.expected.Logging.Output, cfg.Logging.Output)
				return
			}

			assert.Error(t, err)
		})
	}
}

func TestGetConfig_DefaultConfig(t *testing.T) {
	t.Parallel()

	nonExistentFile := "/path/to/nonexistent/file.yaml"
	cfg, err := config.GetConfig(nonExistentFile)
	require.NoError(t, err)

	expected := config.Config{
		Engine: &config.EngineConfig{
			Type: "in_memory",
		},
		Network: &config.NetworkConfig{
			Address:        "127.0.0.1:3223",
			MaxConnections: 100,
			MaxMessageSize: "4KB",
			IdleTimeout:    20 * time.Minute,
		},
		Logging: &config.LoggingConfig{
			Level:  "info",
			Output: "./log/output.log",
		},
	}

	assert.Equal(t, expected.Engine.Type, cfg.Engine.Type)
	assert.Equal(t, expected.Network.Address, cfg.Network.Address)
	assert.Equal(t, expected.Network.MaxConnections, cfg.Network.MaxConnections)
	assert.Equal(t, expected.Network.MaxMessageSize, cfg.Network.MaxMessageSize)
	assert.Equal(t, expected.Network.IdleTimeout, cfg.Network.IdleTimeout)
	assert.Equal(t, expected.Logging.Level, cfg.Logging.Level)
	assert.Equal(t, expected.Logging.Output, cfg.Logging.Output)
}

func TestGetConfig_InvalidFileContent(t *testing.T) {
	t.Parallel()

	content := `invalid yaml content`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	_, err = config.GetConfig(tmpFile.Name())
	assert.Error(t, err)
}
