package database

import (
	"errors"

	"github.com/neekrasov/kvdb/internal/database/command"
	"github.com/neekrasov/kvdb/internal/database/storage"
	"github.com/neekrasov/kvdb/internal/database/storage/models"
	"github.com/neekrasov/kvdb/pkg/config"
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
	Parse(query string) (*command.Command, error)
}

// Engine defines the interface for storing, retrieving, and deleting key-value pairs.
type Storage interface {
	// Set stores a value for a given key.
	Set(key, value string)
	// Get retrieves the value associated with a given key.
	Get(key string) (string, error)
	// Del removes a key and its value from the storage.
	Del(key string) error
}

// NamespacesStorage - An interface for managing namespaces.
type NamespacesStorage interface {
	// Save - Saves a namespace
	Save(namespace string) error
	// Exists - Checks if a namespace exists.
	Exists(namespace string) bool
	// Delete - Deletes a namespace.
	Delete(namespace string) error
	// List - Retrieves a list of all namespaces.
	List() ([]string, error)
	// Append - Adds a namespace to the list of namespaces.
	Append(namespace string) ([]string, error)
}

// UsersStorage - An interface for managing users.
type UsersStorage interface {
	// Authenticate - Authenticates a user by username and password.
	Authenticate(username, password string) (*models.User, error)
	// Create - Creates a new user.
	Create(username, password string) (*models.User, error)
	// Get - Retrieves a user by username.
	Get(username string) (*models.User, error)
	// SaveRaw - Saves a user object directly to storage.
	SaveRaw(user *models.User) error
	// Delete - Deletes a user by username.
	Delete(username string) error
	// AssignRole - Assigns a role to a user.
	AssignRole(username string, role string) error
	// ListUsernames - Retrieves a list of all usernames.
	ListUsernames() ([]string, error)
	// Append - Adds a username to the list of users.
	Append(user string) ([]string, error)
}

// RolesStorage - An interface for managing roles.
type RolesStorage interface {
	// Save - Saves a role.
	Save(role *models.Role) error
	// Get - Retrieves a role by name.
	Get(name string) (*models.Role, error)
	// Delete - Deletes a role by name.
	Delete(name string) error
	// List - Retrieves a list of all roles.
	List() ([]string, error)
	// Append - Adds a role to the list of roles.
	Append(role string) ([]string, error)
}

// CommandHandler used to execute specific actions based on user input.
type CommandHandler struct {
	// Func - The function to execute for the command.
	Func func(*models.User, []string) string
	// AdminOnly - Indicates whether the command can only be executed by an admin.
	AdminOnly bool
}

// SessionStorage - An interface for managing user sessions.
type SessionStorage interface {
	// Create - Creates a new session for a user.
	Create(username string) (string, error)
	// Get - Retrieves a session by token.
	Get(token string) (*storage.Session, error)
	// Delete - Deletes a session by token.
	Delete(token string) error
}

// Database represents the main entry point for parsing and executing commands.
type Database struct {
	parser           Parser
	storage          Storage
	userStorage      UsersStorage
	namespaceStorage NamespacesStorage
	rolesStorage     RolesStorage
	sessions         SessionStorage
	cfg              config.RootConfig
}

// New creates and initializes a new instance of Database.
func New(
	parser Parser, storage Storage,
	userStorage UsersStorage,
	namespaceStorage NamespacesStorage,
	rolesStorage RolesStorage,
	sessions SessionStorage,
	cfg config.RootConfig,
) *Database {
	return &Database{
		parser:           parser,
		storage:          storage,
		userStorage:      userStorage,
		namespaceStorage: namespaceStorage,
		rolesStorage:     rolesStorage,
		sessions:         sessions,
		cfg:              cfg,
	}
}
