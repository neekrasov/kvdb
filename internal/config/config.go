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
		Engine      *EngineConfig      `yaml:"engine" json:"engine" xml:"engine"`
		Network     *NetworkConfig     `yaml:"network" json:"network" xml:"network"`
		Logging     *LoggingConfig     `yaml:"logging" json:"logging" xml:"logging"`
		WAL         *WALConfig         `yaml:"wal" json:"wal" xml:"wal"`
		Root        *RootConfig        `yaml:"root" json:"root" xml:"root"`
		Replication *ReplicationConfig `yaml:"replication" json:"replication" xml:"replication"`

		// -- default optional params
		DefaultRoles      []RoleConfig      `yaml:"default_roles" json:"default_roles" xml:"default_roles"`
		DefaultNamespaces []NamespaceConfig `yaml:"default_namespaces" json:"default_namespaces" xml:"default_namespaces"`
		DefaultUsers      []UserConfig      `yaml:"default_users" json:"default_users" xml:"default_users"`
	}

	RoleConfig struct {
		Name      string `yaml:"name" json:"name" xml:"name"`
		Get       bool   `yaml:"get" json:"get" xml:"get"`
		Set       bool   `yaml:"set" json:"set" xml:"set"`
		Del       bool   `yaml:"del" json:"del" xml:"del"`
		Namespace string `yaml:"namespace" json:"namespace" xml:"namespace"`
	}

	NamespaceConfig struct {
		Name string `yaml:"name" json:"name" xml:"name"`
	}

	UserConfig struct {
		Username string   `yaml:"username" json:"username" xml:"username"`
		Password string   `yaml:"password" json:"password" xml:"password"`
		Roles    []string `yaml:"roles" json:"roles" xml:"roles"`
	}

	EngineConfig struct {
		Type         string `yaml:"type" json:"type" xml:"type"`
		PartitionNum int    `yaml:"partition_num" json:"partition_num" xml:"partition_num"`
	}

	ReplicationConfig struct {
		ReplicaType       string        `yaml:"replica_type" json:"replica_type" xml:"replica_type"`
		MasterAddress     string        `yaml:"master_address" json:"master_address" xml:"master_address"`
		SyncInterval      time.Duration `yaml:"sync_interval" json:"sync_interval" xml:"sync_interval"`
		MaxReplicasNumber int           `yaml:"max_replicas_number"`
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

	WALConfig struct {
		FlushingBatchSize    int           `yaml:"flushing_batch_size" json:"flushing_batch_size" xml:"flushing_batch_size"`
		FlushingBatchTimeout time.Duration `yaml:"flushing_batch_timeout" json:"flushing_batch_timeout" xml:"flushing_batch_timeout"`
		MaxSegmentSize       string        `yaml:"max_segment_size" json:"max_segment_size" xml:"max_segment_size"`
		Compression          string        `yaml:"compression" json:"compression" xml:"compression"`
		DataDir              string        `yaml:"data_directory" json:"data_directory" xml:"data_directory"`
	}

	RootConfig struct {
		Username string `yaml:"username" json:"username" xml:"username"`
		Password string `yaml:"password" json:"password" xml:"password"`
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
		return fmt.Errorf("cant decode yaml config: %w", err)
	}

	return nil
}

func jsonParser(input io.Reader, config *Config) error {
	decoder := json.NewDecoder(input)
	if err := decoder.Decode(config); err != nil {
		return fmt.Errorf("cant decode json config: %w", err)
	}

	return nil
}
