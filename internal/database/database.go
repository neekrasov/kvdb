package database

import (
	"errors"
)

var (
	// ErrKeyNotFound is returned when a key does not exist in the database.
	ErrKeyNotFound = errors.New("key not found")

	// ErrInvalidSyntax is returned when a query has invalid syntax.
	ErrInvalidSyntax = errors.New("invalid syntax")
)

// Parser parses user queries into executable commands.
type Parser interface {
	// Parse converts a query string into a Command or returns an error for invalid syntax.
	Parse(query string) (*Command, error)
}

// Engine defines the interface for storing, retrieving, and deleting key-value pairs.
type Engine interface {
	// Set stores a value for a given key.
	Set(key, value string)
	// Get retrieves the value associated with a given key.
	Get(key string) (string, error)
	// Del removes a key and its value from the storage.
	Del(key string) error
}

// Handler used to execute specific actions based on user input.
type Handler func([]string) string

// Database represents the main entry point for parsing and executing commands.
type Database struct {
	parser Parser // Component responsible for parsing user queries.
	engine Engine // Component responsible for managing the data storage.
}

// New creates and initializes a new instance of Database.
func New(
	parser Parser, // The parser to use for processing queries.
	engine Engine, // The engine to use for storing and managing data.
) *Database {
	return &Database{
		parser: parser,
		engine: engine,
	}
}
