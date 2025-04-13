package compute

import (
	"errors"
	"strings"
)

const (
	AdminHelpText = `
Available commands for admins:

  Operation commands:
    get <key> [ns namespace] - Retrieve the value associated with a key.
    set <key> <value> [ttl duration] [ns namespace] - Store a value for a given key. Example TTL: 10s, 5m, 1h.
    del <key> [ns namespace] - Remove a key and its value from the storage.

  User commands:
	login <username> <password> - Authenticate a user.
	create user <username> <password> - Create a new user.
	get user <username> - Display information about the requested user.
	delete user <username>  - Delete a user.
	assign role <username> <role> - Assign a role to a user.
	divest role <username> <role> - Divest a role from user.
	users - List all usernames.
	me - Display information about the current user.

  Roles commands:
  	get role <role_name> - Display information about the requested role.
  	create role <role_name> <permissions> <namespace> - Create a new role.
    delete role <role_name> - Delete a role.
    roles - List all roles.

  Namespaces commands:
    create ns <namespace> - Create a new namespace.
    delete ns <namespace> - Delete a namespace.
    ns - List all namespaces.
    set ns <namespace> - Set the current namespace for the user.

  Help command:
    help - Display this help message.

  Other commands:
    watch <key> [ns namespace] - Watches the key and returns the value if it has changed.
    stat - Displays database statistics.
`

	UserHelpText = `
Available commands for users:

  Operation commands:
    get <key> [ns namespace] - Retrieve the value associated with a key.
    set <key> <value> [ttl duration] [ns namespace] - Store a value for a given key.
    del <key> [ns namespace] - Remove a key and its value from the storage.

  User commands:
    login <username> <password> - Authenticate a user.
    me - Display information about the current user.

  Namespaces commands:
    ns - List all namespaces.
    set ns <namespace> - Set the current namespace for the user.

  Help command:
    help - Display this help message.

  Other commands:
    watch <key> [ns namespace] - Watches the key and returns the value if it has changed.
`
)

// CommandType - represents the id of a user command.
type CommandID int

const (
	UnknownCommandID CommandID = iota
	SetCommandID
	GetCommandID
	DelCommandID
)

const (
	KeyArg         = "key"
	ValueArg       = "value"
	TTLArg         = "ttl"
	NSArg          = "ns"
	UsernameArg    = "username"
	PasswordArg    = "password"
	RoleArg        = "role"
	RoleNameArg    = "role_name"
	PermissionsArg = "permissions"
	NamespaceArg   = "namespace"
)

var (
	// ErrInvalidCommand - indicates an invalid command or incorrect arguments.
	ErrInvalidCommand = errors.New("invalid command")

	// ErrKeyNotFound - is returned when a key does not exist in the database.
	ErrKeyNotFound = errors.New("key not found")

	// ErrInvalidSyntax - is returned when a query has invalid syntax.
	ErrInvalidSyntax = errors.New("invalid syntax")
)

// CommandType - represents the type of a user command.
type CommandType string

const (
	CommandUNKNOWN CommandType = "unknown"

	// Operation commands
	CommandGET CommandType = "get"
	CommandDEL CommandType = "del"
	CommandSET CommandType = "set"

	// User commands
	CommandAUTH       CommandType = "login"
	CommandGETUSER    CommandType = "get user"
	CommandCREATEUSER CommandType = "create user"
	CommandDELETEUSER CommandType = "delete user"
	CommandUSERS      CommandType = "users"
	CommandSESSIONS   CommandType = "sessions"
	CommandME         CommandType = "me"

	// Roles commands
	CommandGETROLE    CommandType = "get role"
	CommandCREATEROLE CommandType = "create role"
	CommandDELETEROLE CommandType = "delete role"
	CommandROLES      CommandType = "roles"
	CommandASSIGNROLE CommandType = "assign role"
	CommandDIVESTROLE CommandType = "divest role"

	// Namespaces commands
	CommandCREATENAMESPACE CommandType = "create ns"
	CommandDELETENAMESPACE CommandType = "delete ns"
	CommandGETNAMESPACE    CommandType = "get ns"
	CommandNAMESPACES      CommandType = "ns"
	CommandSETNS           CommandType = "set ns"

	// Help command
	CommandHELP CommandType = "help"

	// Watch command
	CommandWATCH CommandType = "watch"

	// Stat command
	CommandSTAT CommandType = "stat"
)

// String - convert CommandType into string/
func (cmd CommandType) String() string {
	return string(cmd)
}

// Make - creates a line containing a command with an arbitrary number of arguments.
func (cmd CommandType) Make(args ...string) string {
	return cmd.String() + " " + strings.Join(args, " ")
}

// Split - split command by space.
func (cmd CommandType) Split(args ...string) []string {
	return strings.Split(cmd.String(), " ")
}

// Command - represents a user command with a type and its arguments.
type Command struct {
	Type CommandType       // The type of the command (GET, SET, DEL).
	Args map[string]string // The arguments associated with the command.
}

type CommandParam struct {
	Required   bool
	Positional bool
	Position   int
}

// NewCommand - creates a new instance of Command.
func NewCommand(commandType CommandType, args map[string]string) (*Command, error) {
	return &Command{Type: commandType, Args: args}, nil
}
