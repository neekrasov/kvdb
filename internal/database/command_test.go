package database_test

import (
	"errors"
	"testing"

	"github.com/neekrasov/kvdb/internal/database"
	"github.com/stretchr/testify/assert"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		commandType database.CommandType
		args        []string
		expectError bool
	}{
		{"Valid GET", database.CommandGET, []string{"key1"}, false},
		{"Valid SET", database.CommandSET, []string{"key1", "value1"}, false},
		{"Valid DEL", database.CommandDEL, []string{"key1"}, false},
		{"Invalid GET (no args)", database.CommandGET, []string{}, true},
		{"Invalid GET (too many args)", database.CommandGET, []string{"key1", "extra"}, true},
		{"Invalid SET (no args)", database.CommandSET, []string{}, true},
		{"Invalid SET (only key)", database.CommandSET, []string{"key1"}, true},
		{"Invalid SET (too many args)", database.CommandSET, []string{"key1", "value1", "extra"}, true},
		{"Invalid DEL (no args)", database.CommandDEL, []string{}, true},
		{"Invalid DEL (too many args)", database.CommandDEL, []string{"key1", "extra"}, true},
		{"Invalid command type", "UNKNOWN", []string{"key1"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd, err := database.NewCommand(tt.commandType, tt.args)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, cmd)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cmd)
				assert.Equal(t, tt.commandType, cmd.Type)
				assert.Equal(t, tt.args, cmd.Args)
			}
		})
	}
}

func TestCommandValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cmd         *database.Command
		expectError bool
	}{
		{"Valid GET", &database.Command{Type: database.CommandGET, Args: []string{"key1"}}, false},
		{"Valid SET", &database.Command{Type: database.CommandSET, Args: []string{"key1", "value1"}}, false},
		{"Valid DEL", &database.Command{Type: database.CommandDEL, Args: []string{"key1"}}, false},
		{"Invalid GET (wrong args)", &database.Command{Type: database.CommandGET, Args: []string{}}, true},
		{"Invalid SET (wrong args)", &database.Command{Type: database.CommandSET, Args: []string{"key1"}}, true},
		{"Invalid DEL (wrong args)", &database.Command{Type: database.CommandDEL, Args: []string{"key1", "extra"}}, true},
		{"Unknown command type", &database.Command{Type: "UNKNOWN", Args: []string{"key1"}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.cmd.Validate()
			if tt.expectError {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, database.ErrInvalidCommand))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
