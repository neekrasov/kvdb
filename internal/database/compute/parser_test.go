package compute

import (
	"fmt"
	"testing"

	"github.com/neekrasov/kvdb/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initCommandTrie() *TrieNode {
	root := NewTrieNode()
	root.Insert(CommandSET, map[string]CommandParam{
		KeyArg:   {Required: true, Positional: true, Position: 0},
		ValueArg: {Required: true, Positional: true, Position: 1},
		TTLArg:   {Required: false, Positional: false},
		NSArg:    {Required: false, Positional: false},
	})
	root.Insert(CommandGET, map[string]CommandParam{
		KeyArg: {Required: true, Positional: true, Position: 0},
		TTLArg: {Required: false, Positional: false},
		NSArg:  {Required: false, Positional: false},
	})
	root.Insert(CommandDEL, map[string]CommandParam{
		KeyArg: {Required: true, Positional: true, Position: 0},
		TTLArg: {Required: false, Positional: false},
		NSArg:  {Required: false, Positional: false},
	})
	root.Insert(CommandAUTH, map[string]CommandParam{
		UsernameArg: {Required: true, Positional: true, Position: 0},
		PasswordArg: {Required: true, Positional: true, Position: 1},
	})
	root.Insert(CommandCREATEUSER, map[string]CommandParam{
		UsernameArg: {Required: true, Positional: true, Position: 0},
		PasswordArg: {Required: true, Positional: true, Position: 1},
	})
	root.Insert(CommandGETUSER, map[string]CommandParam{
		UsernameArg: {Required: true, Positional: true, Position: 0},
	})
	root.Insert(CommandDELETEUSER, map[string]CommandParam{
		UsernameArg: {Required: true, Positional: true, Position: 0},
	})
	root.Insert(CommandASSIGNROLE, map[string]CommandParam{
		UsernameArg: {Required: true, Positional: true, Position: 0},
		RoleArg:     {Required: true, Positional: true, Position: 1},
	})
	root.Insert(CommandDIVESTROLE, map[string]CommandParam{
		UsernameArg: {Required: true, Positional: true, Position: 0},
		RoleArg:     {Required: true, Positional: true, Position: 1},
	})
	root.Insert(CommandCREATEROLE, map[string]CommandParam{
		RoleNameArg:    {Required: true, Positional: true, Position: 0},
		PermissionsArg: {Required: true, Positional: true, Position: 1},
		NamespaceArg:   {Required: true, Positional: true, Position: 2},
	})
	root.Insert(CommandGETROLE, map[string]CommandParam{
		RoleNameArg: {Required: true, Positional: true, Position: 0},
	})
	root.Insert(CommandDELETEROLE, map[string]CommandParam{
		RoleNameArg: {Required: true, Positional: true, Position: 0},
	})
	root.Insert(CommandCREATENAMESPACE, map[string]CommandParam{
		NamespaceArg: {Required: true, Positional: true, Position: 0},
	})
	root.Insert(CommandDELETENAMESPACE, map[string]CommandParam{
		NamespaceArg: {Required: true, Positional: true, Position: 0},
	})
	root.Insert(CommandSETNS, map[string]CommandParam{
		NamespaceArg: {Required: true, Positional: true, Position: 0},
	})
	root.Insert(CommandUSERS, nil)
	root.Insert(CommandME, nil)
	root.Insert(CommandROLES, nil)
	root.Insert(CommandNAMESPACES, nil)
	root.Insert(CommandSESSIONS, nil)
	root.Insert(CommandHELP, nil)
	root.Insert(CommandWATCH, map[string]CommandParam{
		KeyArg: {Required: true, Positional: true, Position: 0},
	})
	root.Insert(CommandSTAT, nil)

	return root
}

func TestParse(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	parser := NewParser(initCommandTrie())

	tests := []struct {
		name        string
		query       string
		expectedCmd *Command
		expectedErr error
	}{
		{
			name:  "Valid SET Query (Positional Only)",
			query: fmt.Sprintf("%s mykey myvalue", CommandSET),
			expectedCmd: &Command{
				Type: CommandSET,
				Args: map[string]string{
					KeyArg:   "mykey",
					ValueArg: "myvalue",
				},
			},
			expectedErr: nil,
		},
		{
			name:  "Valid SET Query (With Optional Named Args)",
			query: fmt.Sprintf("%s mykey myvalue ttl 10s ns testing", CommandSET),
			expectedCmd: &Command{
				Type: CommandSET,
				Args: map[string]string{
					KeyArg:   "mykey",
					ValueArg: "myvalue",
					TTLArg:   "10s",
					NSArg:    "testing",
				},
			},
			expectedErr: nil,
		},
		{
			name:  "Valid SET Query (Optional Named Args Order Swapped)",
			query: fmt.Sprintf("%s mykey myvalue ns testing ttl 5m", CommandSET),
			expectedCmd: &Command{
				Type: CommandSET,
				Args: map[string]string{
					KeyArg:   "mykey",
					ValueArg: "myvalue",
					NSArg:    "testing",
					TTLArg:   "5m",
				},
			},
			expectedErr: nil,
		},
		{
			name:  "Valid GET Query",
			query: fmt.Sprintf("%s somekey", CommandGET),
			expectedCmd: &Command{
				Type: CommandGET,
				Args: map[string]string{
					KeyArg: "somekey",
				},
			},
			expectedErr: nil,
		},
		{
			name:  "Valid Multi-Word Command (CREATE USER)",
			query: fmt.Sprintf("%s newuser securepass", CommandCREATEUSER),
			expectedCmd: &Command{
				Type: CommandCREATEUSER,
				Args: map[string]string{
					UsernameArg: "newuser",
					PasswordArg: "securepass",
				},
			},
			expectedErr: nil,
		},
		{
			name:  "Valid No-Arg Command (USERS)",
			query: CommandUSERS.String(),
			expectedCmd: &Command{
				Type: CommandUSERS,
				Args: map[string]string{},
			},
			expectedErr: nil,
		},
		{
			name:  "Valid No-Arg Command (ME)",
			query: CommandME.String(),
			expectedCmd: &Command{
				Type: CommandME,
				Args: map[string]string{},
			},
			expectedErr: nil,
		},
		{
			name:        "Empty Query",
			query:       "",
			expectedCmd: nil,
			expectedErr: fmt.Errorf("%w: query cannot be empty", ErrInvalidSyntax),
		},
		{
			name:        "Whitespace Query",
			query:       "   ",
			expectedCmd: nil,
			expectedErr: fmt.Errorf("%w: query cannot be empty", ErrInvalidSyntax),
		},
		{
			name:        "Unknown Command",
			query:       "UNKNOWN command value",
			expectedCmd: nil,
			expectedErr: fmt.Errorf("%w: unknown command", ErrInvalidCommand),
		},
		{
			name:        "SET Missing Value Arg",
			query:       fmt.Sprintf("%s keyonly", CommandSET),
			expectedCmd: nil,
			expectedErr: fmt.Errorf("%w: missing required parameter '%s'", ErrInvalidSyntax, ValueArg),
		},
		{
			name:        "GET Missing Key Arg",
			query:       CommandGET.String(),
			expectedCmd: nil,
			expectedErr: fmt.Errorf("%w: missing required parameter '%s'", ErrInvalidSyntax, KeyArg),
		},
		{
			name:        "SET With Missing Value for Named Arg (TTL)",
			query:       fmt.Sprintf("%s key value TTL", CommandSET),
			expectedCmd: nil,
			expectedErr: fmt.Errorf("%w: missing value for parameter '%s'", ErrInvalidSyntax, "TTL"),
		},
		{
			name:        "SET With Unknown Named Arg",
			query:       fmt.Sprintf("%s key value UNKNOWNARG foo", CommandSET),
			expectedCmd: nil,
			expectedErr: fmt.Errorf("%w: unknown parameter '%s'", ErrInvalidSyntax, "UNKNOWNARG"),
		},
		{
			name:        "CREATE USER Missing Password Arg",
			query:       fmt.Sprintf("%s justuser", CommandCREATEUSER),
			expectedCmd: nil,
			expectedErr: fmt.Errorf("%w: missing required parameter '%s'", ErrInvalidSyntax, PasswordArg),
		},
		{
			name:        "Command Prefix of Another (e.g., 'create' instead of 'create user')",
			query:       "create useronly",
			expectedCmd: nil,
			expectedErr: fmt.Errorf("%w: unknown command", ErrInvalidCommand),
		},
		{
			name:        "Named Arg Before Positional Completed",
			query:       fmt.Sprintf("%s key TTL 10s value", CommandSET),
			expectedCmd: nil,
			expectedErr: fmt.Errorf("%w: unknown parameter '%s'", ErrInvalidSyntax, "10s"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd, err := parser.Parse(tt.query)

			if tt.expectedErr != nil {
				require.Error(t, err, "Query: %s", tt.query)
				assert.Equal(t, tt.expectedErr.Error(), err.Error(), "Query: %s", tt.query)
				assert.Nil(t, cmd, "Command should be nil on error. Query: %s", tt.query)
			} else {
				require.NoError(t, err, "Query: %s", tt.query)
				require.NotNil(t, cmd, "Command should not be nil on success. Query: %s", tt.query)
				assert.Equal(t, tt.expectedCmd.Type, cmd.Type, "Query: %s", tt.query)
				assert.Equal(t, tt.expectedCmd.Args, cmd.Args, "Query: %s", tt.query)
			}
		})
	}
}
