package compute

import (
	"fmt"
	"testing"

	"github.com/neekrasov/kvdb/internal/database"
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
		expectedCmd *database.Command
		expectedErr error
	}{
		{
			name:  "Valid Query",
			query: "SET key value",
			expectedCmd: &database.Command{
				Type: "SET",
				Args: []string{"key", "value"},
			},
		},
		{
			name:  "Invalid Query (one token)",
			query: "SET",
			expectedCmd: &database.Command{
				Type: "SET",
				Args: []string{},
			},
			expectedErr: fmt.Errorf("%w: SET command requires exactly 2 arguments", database.ErrInvalidCommand),
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
			expectedErr: fmt.Errorf("%w: unrecognized command 'UNKNOWN'", database.ErrInvalidCommand),
		},
	}

	parser := NewParser()
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
