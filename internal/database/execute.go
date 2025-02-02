package database

import (
	"fmt"
	"strings"

	"github.com/neekrasov/kvdb/pkg/logger"
	"go.uber.org/zap"
)

// wrapError - wrapping error with prefix 'error:'.
func wrapError(err error) string {
	return fmt.Sprintf("error: %s", err.Error())
}

// isError - check the prefix 'error:' exists.
func isError(val string) bool {
	return strings.Contains(val, "error:")
}

// HandleQuery processes a user query by parsing and executing the corresponding command.
func (c *Database) HandleQuery(query string) string {
	cmd, err := c.parser.Parse(query)
	if err != nil {
		return wrapError(fmt.Errorf("parse input failed: %w", err))
	}

	logger.Info("parsed command",
		zap.String("cmd_type", string(cmd.Type)),
		zap.Strings("args", cmd.Args))

	val := map[CommandType]Handler{
		CommandGET: c.get,
		CommandSET: c.set,
		CommandDEL: c.del,
	}[cmd.Type](cmd.Args)

	logger.Info("operation executed",
		zap.String("cmd_type", string(cmd.Type)),
		zap.Strings("args", cmd.Args),
		zap.String("result", val),
		zap.Bool("error", isError(val)))

	return val
}

// del executes the DEL command to remove a key.
func (c *Database) del(args []string) string {
	if err := c.engine.Del(args[0]); err != nil {
		return wrapError(err)
	}

	return args[0]
}

// get executes the GET command to retrieve the value of a key.
func (c *Database) get(args []string) string {
	val, err := c.engine.Get(args[0])
	if err != nil {
		return wrapError(err)
	}

	return val
}

// set executes the SET command to store a key-value pair.
func (c *Database) set(args []string) string {
	c.engine.Set(args[0], args[1])
	return args[1]
}
