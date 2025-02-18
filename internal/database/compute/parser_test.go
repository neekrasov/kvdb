package compute

import (
	"fmt"
	"testing"

	"github.com/neekrasov/kvdb/internal/database"
	"github.com/neekrasov/kvdb/internal/database/models"
	"github.com/neekrasov/kvdb/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_TableDriven(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	tests := []struct {
		name        string
		query       string
		expectedCmd *models.Command
		expectedErr error
	}{
		{
			name:  "Valid Query",
			query: fmt.Sprintf("%s key value", models.CommandSET),
			expectedCmd: &models.Command{
				Type: models.CommandSET,
				Args: []string{"key", "value"},
			},
		},
		{
			name:  "Invalid Query (one token)",
			query: string(models.CommandSET),
			expectedCmd: &models.Command{
				Type: models.CommandSET,
				Args: []string{},
			},
			expectedErr: fmt.Errorf("%w: %s command requires exactly 2 arguments", models.ErrInvalidCommand, models.CommandSET),
		},
		{
			name:        "Empty Query",
			query:       "",
			expectedErr: fmt.Errorf("%w: query cannot be empty", database.ErrInvalidSyntax),
		},
		{
			name:        "Invalid Query (empty)",
			query:       "   ",
			expectedErr: fmt.Errorf("%w: query cannot be empty", database.ErrInvalidSyntax),
		},
		{
			name:        "Invalid Query (unknown command)",
			query:       "UNKNOWN command value",
			expectedErr: fmt.Errorf("%w: unrecognized command", models.ErrInvalidCommand),
		},
	}

	parser := NewParser(initCommandTrie())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := parser.Parse(tt.query)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedCmd.Type, cmd.Type)
				assert.Equal(t, tt.expectedCmd.Args, cmd.Args)
			}
		})
	}
}

func initCommandTrie() *TrieNode {
	root := NewTrieNode()
	root.Insert([]string{"create", "role"}, models.CommandCREATEROLE)
	root.Insert([]string{"create", "user"}, models.CommandCREATEUSER)
	root.Insert([]string{"assign", "role"}, models.CommandASSIGNROLE)
	root.Insert([]string{"delete", "role"}, models.CommandDELETEROLE)
	root.Insert([]string{"create", "ns"}, models.CommandCREATENAMESPACE)
	root.Insert([]string{"delete", "ns"}, models.CommandDELETENAMESPACE)
	root.Insert([]string{"set", "ns"}, models.CommandSETNS)
	root.Insert([]string{"get"}, models.CommandGET)
	root.Insert([]string{"set"}, models.CommandSET)
	root.Insert([]string{"del"}, models.CommandDEL)
	root.Insert([]string{"login"}, models.CommandAUTH)
	root.Insert([]string{"users"}, models.CommandUSERS)
	root.Insert([]string{"me"}, models.CommandME)
	root.Insert([]string{"roles"}, models.CommandROLES)
	root.Insert([]string{"ns"}, models.CommandNAMESPACES)
	root.Insert([]string{"help"}, models.CommandHELP)

	return root
}
