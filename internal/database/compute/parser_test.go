package compute

import (
	"fmt"
	"testing"

	"github.com/neekrasov/kvdb/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
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
			expectedErr: fmt.Errorf("%w: %s command requires 2 or 3 arguments", ErrInvalidCommand, CommandSET),
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
			t.Parallel()

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
	root.Insert(CommandSET, "set")

	return root
}
