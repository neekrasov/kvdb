package compute

import (
	"errors"
	"fmt"
	"strings"

	"github.com/neekrasov/kvdb/pkg/logger"
	"go.uber.org/zap"
)

const (
	AdminHelpText = `
Available commands for admins:

  Operation commands:
    get <key> - Retrieve the value associated with a key.
    set <key> <value> [ttl] - Store a value for a given key.
    del <key> - Remove a key and its value from the storage.

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
`

	UserHelpText = `
Available commands for users:

  Operation commands:
    get <key> - Retrieve the value associated with a key.
    set <key> <value> [ttl] - Store a value for a given key.
    del <key> - Remove a key and its value from the storage.

  User commands:
    login <username> <password> - Authenticate a user.
    me - Display information about the current user.

  Namespaces commands:
    ns - List all namespaces.
    set ns <namespace> - Set the current namespace for the user.

  Help command:
    help - Display this help message.
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

// Command - represents a user command with a type and its arguments.
type Command struct {
	Type CommandType // The type of the command (GET, SET, DEL).
	Args []string    // The arguments associated with the command.
}

// NewCommand - creates a new instance of Command and validates it.
func NewCommand(commandType CommandType, args []string) (*Command, error) {
	zapargs := []zap.Field{
		zap.Stringer("cmd_type", commandType),
		zap.Strings("args", args),
	}

	cmd := &Command{Type: commandType, Args: args}
	if err := cmd.Validate(); err != nil {
		zapargs = append(zapargs, zap.Error(err))
		logger.Debug("invalid command", zapargs...)
		return nil, err
	}

	logger.Debug("command successfully created", zapargs...)

	return cmd, nil
}

// Validate - checks if the command type and arguments are valid.
func (cmd *Command) Validate() error {
	switch cmd.Type {
	case CommandGET, CommandDEL, CommandDELETEROLE,
		CommandCREATENAMESPACE, CommandDELETENAMESPACE,
		CommandSETNS, CommandDELETEUSER, CommandGETUSER,
		CommandGETROLE, CommandWATCH:
		if len(cmd.Args) != 1 {
			return fmt.Errorf("%w: %s command requires exactly 1 argument", ErrInvalidCommand, cmd.Type)
		}
	case CommandAUTH, CommandASSIGNROLE,
		CommandCREATEUSER, CommandDIVESTROLE:
		if len(cmd.Args) != 2 {
			return fmt.Errorf("%w: %s command requires exactly 2 arguments", ErrInvalidCommand, cmd.Type)
		}
	case CommandSET: // (key, value, [ttl])
		if len(cmd.Args) < 2 || len(cmd.Args) > 3 {
			return fmt.Errorf("%w: %s command requires 2 or 3 arguments", ErrInvalidCommand, cmd.Type)
		}
	case CommandCREATEROLE:
		if len(cmd.Args) != 3 {
			return fmt.Errorf("%w: %s command requires exactly 3 arguments", ErrInvalidCommand, cmd.Type)
		}
	case CommandUSERS, CommandME, CommandROLES,
		CommandNAMESPACES, CommandHELP, CommandSESSIONS,
		CommandSTAT:
		if len(cmd.Args) > 0 {
			return fmt.Errorf("%w: command '%s' does not accept arguments", ErrInvalidCommand, cmd.Type)
		}
	default:
		return fmt.Errorf("%w: unrecognized command", ErrInvalidCommand)
	}

	return nil
}
