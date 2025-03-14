package compute_test

import (
	"testing"

	"github.com/neekrasov/kvdb/internal/database/compute"
	"github.com/neekrasov/kvdb/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	tests := []struct {
		name        string
		commandType compute.CommandType
		args        []string
		expectError bool
	}{
		{"Valid GET", compute.CommandGET, []string{"key1"}, false},
		{"Valid SET", compute.CommandSET, []string{"key1", "value1"}, false},
		{"Valid DEL", compute.CommandDEL, []string{"key1"}, false},
		{"Invalid GET (no args)", compute.CommandGET, []string{}, true},
		{"Invalid GET (too many args)", compute.CommandGET, []string{"key1", "extra"}, true},
		{"Invalid Me (too many args)", compute.CommandME, []string{"extra"}, true},
		{"Invalid  (too many args)", compute.CommandME, []string{"extra"}, true},
		{"Invalid SET (no args)", compute.CommandSET, []string{}, true},
		{"Invalid SET (only key)", compute.CommandSET, []string{"key1"}, true},
		{"Invalid CREATE ROLE (only key)", compute.CommandCREATEROLE, []string{"key1", "key2"}, true},
		{"Invalid SET (too many args)", compute.CommandSET, []string{"key1", "value1", "1s", "extra"}, true},
		{"Invalid DEL (no args)", compute.CommandDEL, []string{}, true},
		{"Invalid DEL (too many args)", compute.CommandDEL, []string{"key1", "extra"}, true},
		{"Invalid compute type", "UNKNOWN", []string{"key1"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd, err := compute.NewCommand(tt.commandType, tt.args)
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
