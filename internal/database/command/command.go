package command

import (
	"errors"
	"fmt"

	"github.com/neekrasov/kvdb/pkg/logger"
	"go.uber.org/zap"
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
	// Operation commands
	CommandGET CommandType = "get" // Command to retrieve a value by key.
	CommandDEL CommandType = "del" // Command to delete a key.
	CommandSET CommandType = "set" // Command to set a value for a key.

	// User commands
	CommandAUTH       CommandType = "login"
	CommandCREATEUSER CommandType = "create_user"
	CommandASSIGNROLE CommandType = "assign_role"
	CommandUSERS      CommandType = "users"
	CommandME         CommandType = "me"

	// Roles commands
	CommandCREATEROLE CommandType = "create_role"
	CommandDELETEROLE CommandType = "delete_role"
	CommandROLES      CommandType = "roles"

	// Namespaces commands
	CommandCREATENAMESPACE CommandType = "create_ns"
	CommandDELETENAMESPACE CommandType = "delete_ns"
	CommandNAMESPACES      CommandType = "ns"
	CommandSETNS           CommandType = "set_ns"
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
	case CommandUSERS, CommandME, CommandROLES, CommandNAMESPACES:
		if len(cmd.Args) > 0 {
			return fmt.Errorf("%w: command '%s' does not accept arguments", ErrInvalidCommand, cmd.Type)
		}
	default:
		return fmt.Errorf("%w: unrecognized command '%s'", ErrInvalidCommand, cmd.Type)
	}

	return nil
}
