package database

import (
	"errors"
	"fmt"

	"github.com/neekrasov/kvdb/pkg/logger"
	"go.uber.org/zap"
)

// ErrInvalidCommand indicates an invalid command or incorrect arguments.
var ErrInvalidCommand = errors.New("invalid command")

// CommandType represents the type of a user command.
type CommandType string

const (
	CommandGET CommandType = "GET" // Command to retrieve a value by key.
	CommandDEL CommandType = "DEL" // Command to delete a key.
	CommandSET CommandType = "SET" // Command to set a value for a key.
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
	zapargs := []zap.Field{zap.Stringer("cmd_type", commandType),
		zap.Strings("args", args)}

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
	case CommandGET, CommandDEL:
		if len(cmd.Args) != 1 {
			return fmt.Errorf("%w: %s command requires exactly 1 argument", ErrInvalidCommand, cmd.Type)
		}
	case CommandSET:
		if len(cmd.Args) != 2 {
			return fmt.Errorf("%w: %s command requires exactly 2 arguments", ErrInvalidCommand, cmd.Type)
		}
	default:
		return fmt.Errorf("%w: unrecognized command '%s'", ErrInvalidCommand, cmd.Type)
	}

	return nil
}
