package database

import (
	"context"

	"github.com/neekrasov/kvdb/internal/config"
	"github.com/neekrasov/kvdb/internal/database/compute"
	"github.com/neekrasov/kvdb/internal/database/identity/models"
	"github.com/neekrasov/kvdb/internal/database/storage"
	pkgsync "github.com/neekrasov/kvdb/pkg/sync"
)

// Stats - structure for storing statistical databases.
type Stats struct {
	Uptime          float64 `json:"uptime"`           // Server uptime.
	TotalCommands   int64   `json:"total_commands"`   // Total number of commands executed.
	GetCommands     int64   `json:"get_commands"`     // Number of GET commands.
	SetCommands     int64   `json:"set_commands"`     // Number of SET commands.
	DelCommands     int64   `json:"del_commands"`     // Number of DEL commands.
	TotalKeys       int64   `json:"total_keys"`       // Total number of keys in the storage (approximate).
	ExpiredKeys     int64   `json:"expired_keys"`     // Number of expired keys (deleted).
	ActiveSessions  int64   `json:"active_sessions"`  // Number of active sessions.
	TotalNamespaces int64   `json:"total_namespaces"` // Number of namespaces.
	TotalRoles      int64   `json:"total_roles"`      // Number of roles.
	TotalUsers      int64   `json:"total_users"`      // Number of users.
}

// Parser - parses user queries into executable commands.
type Parser interface {
	// Parse - converts a query string into a Command or returns an error for invalid syntax.
	Parse(query string) (*compute.Command, error)
}

// Storage - interface for storing, retrieving, and deleting key-value pairs.
type Storage interface {
	// Set - stores a value for a given key.
	Set(ctx context.Context, key, value string) error
	// Get - retrieves the value associated with a given key.
	Get(ctx context.Context, key string) (string, error)
	// Del - removes a key and its value from the storage.
	Del(ctx context.Context, key string) error
	// Watch - watches the key and returns the value if it has changed.
	Watch(ctx context.Context, key string) pkgsync.FutureString
	// Stats - returns the collected database statistics.
	Stats() (*storage.Stats, error)
}

// NamespacesStorage - interface for managing namespaces.
type NamespacesStorage interface {
	// Save - Saves a namespace
	Save(ctx context.Context, namespace string) error
	// Exists - Checks if a namespace exists.
	Exists(ctx context.Context, namespace string) bool
	// Delete - Deletes a namespace.
	Delete(ctx context.Context, namespace string) error
	// List - Retrieves a list of all namespaces.
	List(ctx context.Context) ([]string, error)
	// Append - Adds a namespace to the list of namespaces.
	Append(ctx context.Context, namespace string) ([]string, error)
}

// UsersStorage - interface for managing users.
type UsersStorage interface {
	// Authenticate - authenticates a user by username and password.
	Authenticate(ctx context.Context, username, password string) (*models.User, error)
	// Create - creates a new user.
	Create(ctx context.Context, username, password string) (*models.User, error)
	// Get - retrieves a user by username.
	Get(ctx context.Context, username string) (*models.User, error)
	// SaveRaw - saves a user object directly to storage.
	SaveRaw(ctx context.Context, user *models.User) error
	// Delete - deletes a user by username.
	Delete(ctx context.Context, username string) error
	// AssignRole - assigns a role to a user.
	AssignRole(ctx context.Context, username string, role string) error
	// ListUsernames - retrieves a list of all usernames.
	ListUsernames(ctx context.Context) ([]string, error)
	// Append - adds a username to the list of users.
	Append(ctx context.Context, username string) ([]string, error)
	// Remove - remove username from the list of all users in the system.
	Remove(ctx context.Context, username string) ([]string, error)
}

// RolesStorage - interface for managing roles.
type RolesStorage interface {
	// Save - saves a role.
	Save(ctx context.Context, role *models.Role) error
	// Get - retrieves a role by name.
	Get(ctx context.Context, name string) (*models.Role, error)
	// Delete - deletes a role by name.
	Delete(ctx context.Context, name string) error
	// List - retrieves a list of all roles.
	List(ctx context.Context) ([]string, error)
	// Append - adds a role to the list of roles.
	Append(ctx context.Context, role string) ([]string, error)
}

// CommandHandler - used to execute specific actions based on user input.
type CommandHandler struct {
	// Func - function to execute for the models.
	Func func(context.Context, *models.User, []string) string
	// AdminOnly - indicates whether the models can only be executed by an admin.
	AdminOnly bool
}

// SessionStorage - interface for managing user sessions.
type SessionStorage interface {
	// Create - creates a new session for a user.
	Create(id string, user *models.User) error
	// Get - retrieves a session by token.
	Get(id string) (*models.Session, error)
	// Delete - deletes a session by token.
	Delete(id string)
	// List - retrieves a list of all active sessions.
	List() []models.Session
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
	registry         map[compute.CommandType]CommandHandler
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
	db := Database{
		parser:           parser,
		storage:          storage,
		userStorage:      userStorage,
		namespaceStorage: namespaceStorage,
		rolesStorage:     rolesStorage,
		sessions:         sessions,
		cfg:              cfg,
	}

	db.registry = map[compute.CommandType]CommandHandler{
		compute.CommandCREATEUSER:      {Func: db.createUser, AdminOnly: true},
		compute.CommandASSIGNROLE:      {Func: db.assignRole, AdminOnly: true},
		compute.CommandCREATEROLE:      {Func: db.createRole, AdminOnly: true},
		compute.CommandDELETEROLE:      {Func: db.delRole, AdminOnly: true},
		compute.CommandROLES:           {Func: db.listRoles, AdminOnly: true},
		compute.CommandGETROLE:         {Func: db.getRole, AdminOnly: true},
		compute.CommandUSERS:           {Func: db.users, AdminOnly: true},
		compute.CommandGETUSER:         {Func: db.getUser, AdminOnly: true},
		compute.CommandCREATENAMESPACE: {Func: db.createNS, AdminOnly: true},
		compute.CommandDELETENAMESPACE: {Func: db.deleteNS, AdminOnly: true},
		compute.CommandNAMESPACES:      {Func: db.ns, AdminOnly: true},
		compute.CommandSESSIONS:        {Func: db.listSessions, AdminOnly: true},
		compute.CommandDELETEUSER:      {Func: db.deleteUser, AdminOnly: true},
		compute.CommandDIVESTROLE:      {Func: db.divestRole, AdminOnly: true},
		compute.CommandSTAT:            {Func: db.stat, AdminOnly: true},
		compute.CommandHELP:            {Func: db.help},
		compute.CommandSETNS:           {Func: db.setNamespace},
		compute.CommandME:              {Func: db.me},
		compute.CommandGET:             {Func: db.get},
		compute.CommandSET:             {Func: db.set},
		compute.CommandDEL:             {Func: db.del},
		compute.CommandWATCH:           {Func: db.watch},
	}

	return &db
}
