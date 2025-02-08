package command_test

import (
	"testing"

	"github.com/neekrasov/kvdb/internal/database/command"
	"github.com/neekrasov/kvdb/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	tests := []struct {
		name        string
		commandType command.CommandType
		args        []string
		expectError bool
	}{
		{"Valid GET", command.CommandGET, []string{"key1"}, false},
		{"Valid SET", command.CommandSET, []string{"key1", "value1"}, false},
		{"Valid DEL", command.CommandDEL, []string{"key1"}, false},
		{"Invalid GET (no args)", command.CommandGET, []string{}, true},
		{"Invalid GET (too many args)", command.CommandGET, []string{"key1", "extra"}, true},
		{"Invalid Me (too many args)", command.CommandME, []string{"extra"}, true},
		{"Invalid  (too many args)", command.CommandME, []string{"extra"}, true},
		{"Invalid SET (no args)", command.CommandSET, []string{}, true},
		{"Invalid SET (only key)", command.CommandSET, []string{"key1"}, true},
		{"Invalid CREATE ROLE (only key)", command.CommandCREATEROLE, []string{"key1", "key2"}, true},
		{"Invalid SET (too many args)", command.CommandSET, []string{"key1", "value1", "extra"}, true},
		{"Invalid DEL (no args)", command.CommandDEL, []string{}, true},
		{"Invalid DEL (too many args)", command.CommandDEL, []string{"key1", "extra"}, true},
		{"Invalid command type", "UNKNOWN", []string{"key1"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd, err := command.NewCommand(tt.commandType, tt.args)
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
