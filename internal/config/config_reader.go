package config

import (
	"bytes"
	"io"
	"os"
)

func GetConfigReader(path string) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if err == nil {
		return f, nil
	}

	var defaultConfigYaml = `engine:
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

	var bb bytes.Buffer
	if _, err = bb.WriteString(defaultConfigYaml); err != nil {
		return nil, err
	}

	return io.NopCloser(&bb), nil
}
