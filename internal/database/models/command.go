package models

import (
	"errors"
	"fmt"

	"github.com/neekrasov/kvdb/pkg/logger"
	"go.uber.org/zap"
)

const (
	AdminHelpText = `
Available commands for admins:

  Operation commands:
    get <key>            - Retrieve the value associated with a key.
    set <key> <value>    - Store a value for a given key.
    del <key>            - Remove a key and its value from the storage.

  User commands:
    login <username> <password> - Authenticate a user.
    create_user <username> <password> - Create a new user.
    assign_role <username> <role> - Assign a role to a user.
    users - List all usernames.
    me - Display information about the current user.

  Roles commands:
    create_role <role_name> <permissions> <namespace> - Create a new role.
    delete_role <role_name> - Delete a role.
    roles - List all roles.

  Namespaces commands:
    create_ns <namespace> - Create a new namespace.
    delete_ns <namespace> - Delete a namespace.
    ns - List all namespaces.
    set_ns <namespace> - Set the current namespace for the user.

  Help command:
    help - Display this help message.
`

	UserHelpText = `
Available commands for users:

  Operation commands:
    get <key>            - Retrieve the value associated with a key.
    set <key> <value>    - Store a value for a given key.
    del <key>            - Remove a key and its value from the storage.

  User commands:
    login <username> <password> - Authenticate a user.
    me - Display information about the current user.

  Namespaces commands:
    ns - List all namespaces.
    set_ns <namespace> - Set the current namespace for the user.

  Help command:
    help - Display this help message.
`
)

// CommandType represents the id of a user command.
type CommandID int

const (
	UnknownCommandID CommandID = iota
	SetCommandID
	GetCommandID
	DelCommandID
)

// ErrInvalidCommand indicates an invalid command or incorrect arguments.
var ErrInvalidCommand = errors.New("invalid command")

// CommandType represents the type of a user command.
type CommandType string

const (
	CommandUNKNOWN CommandType = "unknown"

	// Operation commands
	CommandGET CommandType = "get"
	CommandDEL CommandType = "del"
	CommandSET CommandType = "set"

	// User commands
	CommandAUTH       CommandType = "login"
	CommandCREATEUSER CommandType = "create user"
	CommandASSIGNROLE CommandType = "assign role"
	CommandUSERS      CommandType = "users"
	CommandME         CommandType = "me"

	// Roles commands
	CommandCREATEROLE CommandType = "create role"
	CommandDELETEROLE CommandType = "delete role"
	CommandROLES      CommandType = "roles"

	// Namespaces commands
	CommandCREATENAMESPACE CommandType = "create ns"
	CommandDELETENAMESPACE CommandType = "delete ns"
	CommandNAMESPACES      CommandType = "ns"
	CommandSETNS           CommandType = "set ns"

	// Help command
	CommandHELP CommandType = "help"
)

// String convert CommandType into string/
func (cmd CommandType) String() string {
	return string(cmd)
}

// Command represents a user command with a type and its arguments.
type Command struct {
	Type CommandType // The type of the command (GET, SET, DEL).
	Args []string    // The arguments associated with the command.
}

// NewCommand creates a new instance of Command and validates it.
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

// Validate checks if the command type and arguments are valid.
func (cmd *Command) Validate() error {
	switch cmd.Type {
	case CommandGET, CommandDEL, CommandDELETEROLE,
		CommandCREATENAMESPACE, CommandDELETENAMESPACE, CommandSETNS:
		if len(cmd.Args) != 1 {
			return fmt.Errorf("%w: %s command requires exactly 1 argument", ErrInvalidCommand, cmd.Type)
		}
	case CommandAUTH, CommandASSIGNROLE, CommandSET, CommandCREATEUSER:
		if len(cmd.Args) != 2 {
			return fmt.Errorf("%w: %s command requires exactly 2 arguments", ErrInvalidCommand, cmd.Type)
		}
	case CommandCREATEROLE:
		if len(cmd.Args) != 3 {
			return fmt.Errorf("%w: %s command requires exactly 3 arguments", ErrInvalidCommand, cmd.Type)
		}
	case CommandUSERS, CommandME, CommandROLES, CommandNAMESPACES, CommandHELP:
		if len(cmd.Args) > 0 {
			return fmt.Errorf("%w: command '%s' does not accept arguments", ErrInvalidCommand, cmd.Type)
		}
	default:
		return fmt.Errorf("%w: unrecognized command", ErrInvalidCommand)
	}

	return nil
}
