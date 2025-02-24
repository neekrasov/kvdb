package compute

import (
	"fmt"
	"testing"

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
		expectedCmd *Command
		expectedErr error
	}{
		{
			name:  "Valid Query",
			query: fmt.Sprintf("%s key value", CommandSET),
			expectedCmd: &Command{
				Type: CommandSET,
				Args: []string{"key", "value"},
			},
		},
		{
			name:  "Invalid Query (one token)",
			query: string(CommandSET),
			expectedCmd: &Command{
				Type: CommandSET,
				Args: []string{},
			},
			expectedErr: fmt.Errorf("%w: %s command requires exactly 2 arguments", ErrInvalidCommand, CommandSET),
		},
		{
			name:        "Empty Query",
			query:       "",
			expectedErr: fmt.Errorf("%w: query cannot be empty", ErrInvalidSyntax),
		},
		{
			name:        "Invalid Query (empty)",
			query:       "   ",
			expectedErr: fmt.Errorf("%w: query cannot be empty", ErrInvalidSyntax),
		},
		{
			name:        "Invalid Query (unknown command)",
			query:       "UNKNOWN command value",
			expectedErr: fmt.Errorf("%w: unrecognized command", ErrInvalidCommand),
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
	root.Insert([]string{"create", "role"}, CommandCREATEROLE)
	root.Insert([]string{"create", "user"}, CommandCREATEUSER)
	root.Insert([]string{"assign", "role"}, CommandASSIGNROLE)
	root.Insert([]string{"delete", "role"}, CommandDELETEROLE)
	root.Insert([]string{"create", "ns"}, CommandCREATENAMESPACE)
	root.Insert([]string{"delete", "ns"}, CommandDELETENAMESPACE)
	root.Insert([]string{"set", "ns"}, CommandSETNS)
	root.Insert([]string{"get"}, CommandGET)
	root.Insert([]string{"set"}, CommandSET)
	root.Insert([]string{"del"}, CommandDEL)
	root.Insert([]string{"login"}, CommandAUTH)
	root.Insert([]string{"users"}, CommandUSERS)
	root.Insert([]string{"me"}, CommandME)
	root.Insert([]string{"roles"}, CommandROLES)
	root.Insert([]string{"ns"}, CommandNAMESPACES)
	root.Insert([]string{"help"}, CommandHELP)

	return root
}
