package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		Engine  EngineConfig  `yaml:"engine" json:"engine" xml:"engine"`
		Network NetworkConfig `yaml:"network" json:"network" xml:"network"`
		Logging LoggingConfig `yaml:"logging" json:"logging" xml:"logging"`
	}

	EngineConfig struct {
		Type string `yaml:"type" json:"type" xml:"type"`
	}

	NetworkConfig struct {
		Address        string        `yaml:"address" json:"address" xml:"address"`
		MaxConnections uint          `yaml:"max_connections" json:"max_connections" xml:"max_connections"`
		MaxMessageSize string        `yaml:"max_message_size" json:"max_message_size" xml:"max_message_size"`
		IdleTimeout    time.Duration `yaml:"idle_timeout" json:"idle_timeout" xml:"idle_timeout"`
	}

	LoggingConfig struct {
		Level  string `yaml:"level" json:"level" xml:"level"`
		Output string `yaml:"output" json:"output" xml:"output"`
	}
)

func GetConfig(path string) (Config, error) {
	configContent, err := GetConfigReader(path)
	if err != nil {
		return Config{}, err
	}

	return ParseConfig(configContent)
}

func ParseConfig(input io.ReadCloser) (Config, error) {
	defer input.Close()

	var (
		cfg      Config
		parseErr strings.Builder
	)

	for _, parser := range []func(io.Reader, *Config) error{yamlParser, jsonParser} {
		var err error
		if err = parser(input, &cfg); err == nil {
			return cfg, nil
		}
		_, _ = parseErr.WriteString(fmt.Sprintf("Error parsing config: %s\n", err.Error()))
	}

	return cfg, errors.New(parseErr.String())
}

func yamlParser(input io.Reader, config *Config) error {
	decoder := yaml.NewDecoder(input)
	if err := decoder.Decode(config); err != nil {
		return fmt.Errorf("cant decode config. Invalid YAML: %w", err)
	}

	return nil
}

func jsonParser(input io.Reader, config *Config) error {
	decoder := json.NewDecoder(input)
	if err := decoder.Decode(config); err != nil {
		return fmt.Errorf("cant decode config. Invalid JSON: %w", err)
	}

	return nil
}
