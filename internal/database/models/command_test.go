package models_test

import (
	"testing"

	"github.com/neekrasov/kvdb/internal/database/models"
	"github.com/neekrasov/kvdb/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	tests := []struct {
		name        string
		commandType models.CommandType
		args        []string
		expectError bool
	}{
		{"Valid GET", models.CommandGET, []string{"key1"}, false},
		{"Valid SET", models.CommandSET, []string{"key1", "value1"}, false},
		{"Valid DEL", models.CommandDEL, []string{"key1"}, false},
		{"Invalid GET (no args)", models.CommandGET, []string{}, true},
		{"Invalid GET (too many args)", models.CommandGET, []string{"key1", "extra"}, true},
		{"Invalid Me (too many args)", models.CommandME, []string{"extra"}, true},
		{"Invalid  (too many args)", models.CommandME, []string{"extra"}, true},
		{"Invalid SET (no args)", models.CommandSET, []string{}, true},
		{"Invalid SET (only key)", models.CommandSET, []string{"key1"}, true},
		{"Invalid CREATE ROLE (only key)", models.CommandCREATEROLE, []string{"key1", "key2"}, true},
		{"Invalid SET (too many args)", models.CommandSET, []string{"key1", "value1", "extra"}, true},
		{"Invalid DEL (no args)", models.CommandDEL, []string{}, true},
		{"Invalid DEL (too many args)", models.CommandDEL, []string{"key1", "extra"}, true},
		{"Invalid models type", "UNKNOWN", []string{"key1"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd, err := models.NewCommand(tt.commandType, tt.args)
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
