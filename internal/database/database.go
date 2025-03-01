package database

import (
	"github.com/neekrasov/kvdb/internal/config"
	"github.com/neekrasov/kvdb/internal/database/compute"
	"github.com/neekrasov/kvdb/internal/database/identity"
	"github.com/neekrasov/kvdb/internal/database/identity/models"
)

// Parser - parses user queries into executable commands.
type Parser interface {
	// Parse - converts a query string into a Command or returns an error for invalid syntax.
	Parse(query string) (*compute.Command, error)
}

// Storage - interface for storing, retrieving, and deleting key-value pairs.
type Storage interface {
	// Set stores a value for a given key.
	Set(key, value string) error
	// Get retrieves the value associated with a given key.
	Get(key string) (string, error)
	// Del removes a key and its value from the storage.
	Del(key string) error
}

// NamespacesStorage - interface for managing namespaces.
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

// UsersStorage - interface for managing users.
type UsersStorage interface {
	// Authenticate - authenticates a user by username and password.
	Authenticate(username, password string) (*models.User, error)
	// Create - creates a new user.
	Create(username, password string) (*models.User, error)
	// Get - retrieves a user by username.
	Get(username string) (*models.User, error)
	// SaveRaw - saves a user object directly to storage.
	SaveRaw(user *models.User) error
	// Delete - deletes a user by username.
	Delete(username string) error
	// AssignRole - assigns a role to a user.
	AssignRole(username string, role string) error
	// ListUsernames - retrieves a list of all usernames.
	ListUsernames() ([]string, error)
	// Append - adds a username to the list of users.
	Append(user string) ([]string, error)
}

// RolesStorage - interface for managing roles.
type RolesStorage interface {
	// Save - saves a role.
	Save(role *models.Role) error
	// Get - retrieves a role by name.
	Get(name string) (*models.Role, error)
	// Delete - deletes a role by name.
	Delete(name string) error
	// List - retrieves a list of all roles.
	List() ([]string, error)
	// Append - adds a role to the list of roles.
	Append(role string) ([]string, error)
}

// CommandHandler - used to execute specific actions based on user input.
type CommandHandler struct {
	// Func - function to execute for the models.
	Func func(*models.User, []string) string
	// AdminOnly - indicates whether the models can only be executed by an admin.
	AdminOnly bool
}

// SessionStorage - interface for managing user sessions.
type SessionStorage interface {
	// Create - creates a new session for a user.
	Create(username string) (string, error)
	// Get - retrieves a session by token.
	Get(token string) (*identity.Session, error)
	// Delete - deletes a session by token.
	Delete(token string) error
}

// Database - represents the main entry point for parsing and executing commands.
type Database struct {
	parser           Parser
	storage          Storage
	userStorage      UsersStorage
	namespaceStorage NamespacesStorage
	rolesStorage     RolesStorage
	sessions         SessionStorage
	cfg              *config.RootConfig
}

// New - creates and initializes a new instance of Database.
func New(
	parser Parser, storage Storage,
	userStorage UsersStorage,
	namespaceStorage NamespacesStorage,
	rolesStorage RolesStorage,
	sessions SessionStorage,
	cfg *config.RootConfig,
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
