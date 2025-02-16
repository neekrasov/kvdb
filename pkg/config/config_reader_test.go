package config_test

import (
	"io"
	"os"
	"testing"

	"github.com/neekrasov/kvdb/pkg/config"
)

func TestGetConfigReader_FileExists(t *testing.T) {
	t.Parallel()

	tmpFile, err := os.CreateTemp("", "config_test.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	testContent := "test: value"
	if _, err := tmpFile.WriteString(testContent); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	reader, err := config.GetConfigReader(tmpFile.Name())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read file content: %v", err)
	}

	if string(data) != testContent {
		t.Errorf("expected content '%s', got '%s'", testContent, string(data))
	}
}

func TestGetConfigReader_FileNotExists(t *testing.T) {
	t.Parallel()

	reader, err := config.GetConfigReader("nonexistent.yaml")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read default config: %v", err)
	}

	expectedContent := `engine:
  type: "in_memory"
network:
  address: "127.0.0.1:3223"
  max_connections: 100
  max_message_size: "4KB"
  idle_timeout: 20m
logging:
  level: "info"
  output: "./log/output.log"
`

	if string(data) != expectedContent {
		t.Errorf("expected default config, got:\n%s", string(data))
	}
}
