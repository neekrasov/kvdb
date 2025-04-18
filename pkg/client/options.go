package client

import (
	"time"

	"github.com/neekrasov/kvdb/internal/database/compression"
)

// callOptions - внутренняя структура для хранения настроек конкретного вызова
type callOptions struct {
	compressor compression.Compressor
	ttl        *time.Duration
	namespace  string
}

// Option - общий тип для опций методов клиента.
type Option func(*callOptions)

// WithCompressor - опция, указывающая, что нужно использовать сжатие для данного вызова.
// Предварительно инициализированный компрессор игнорируется.
func WithCompressor(compressor compression.Compressor) Option {
	return func(o *callOptions) {
		o.compressor = compressor
	}
}

// WithTTL - опция для установки времени жизни ключа (только для Set).
func WithTTL(duration time.Duration) Option {
	return func(o *callOptions) {
		o.ttl = &duration
	}
}

// WithNamespace - опция для указания пространства имен для операции.
// Предварительно инициализированное пространство имён игнорируется.
func WithNamespace(namespace string) Option {
	return func(o *callOptions) {
		o.namespace = namespace
	}
}

func applyOptions(opts []Option) callOptions {
	co := callOptions{}
	for _, opt := range opts {
		opt(&co)
	}
	return co
}
