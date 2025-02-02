package database_test

import (
	"testing"

	"github.com/neekrasov/kvdb/internal/database"
	mocks "github.com/neekrasov/kvdb/internal/mocks/database"
	"github.com/neekrasov/kvdb/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestHandleQuery(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	tests := []struct {
		name        string
		query       string
		parser      func() *mocks.Parser
		engine      func() *mocks.Engine
		expectedRes string
	}{
		{
			name:  "Handle GET Command Successfully",
			query: "GET key1",
			parser: func() *mocks.Parser {
				parser := new(mocks.Parser)
				parser.On("Parse", "GET key1").Return(&database.Command{Type: database.CommandGET, Args: []string{"key1"}}, nil)
				return parser
			},
			engine: func() *mocks.Engine {
				engine := new(mocks.Engine)
				engine.On("Get", "key1").Return("value1", nil)
				return engine
			},
			expectedRes: "value1",
		},
		{
			name:  "Handle SET Command Successfully",
			query: "SET key1 value1",
			parser: func() *mocks.Parser {
				parser := new(mocks.Parser)
				parser.On("Parse", "SET key1 value1").Return(&database.Command{Type: database.CommandSET, Args: []string{"key1", "value1"}}, nil)
				return parser
			},
			engine: func() *mocks.Engine {
				engine := new(mocks.Engine)
				engine.On("Set", "key1", "value1").Return(nil)
				return engine
			},
			expectedRes: "value1",
		},
		{
			name:  "Handle DEL Command Successfully",
			query: "DEL key1",
			parser: func() *mocks.Parser {
				parser := new(mocks.Parser)
				parser.On("Parse", "DEL key1").Return(&database.Command{Type: database.CommandDEL, Args: []string{"key1"}}, nil)
				return parser
			},
			engine: func() *mocks.Engine {
				engine := new(mocks.Engine)
				engine.On("Del", "key1").Return(nil)
				return engine
			},
			expectedRes: "key1",
		},
		{
			name:  "Handle Invalid Command Type",
			query: "INVALID key1",
			parser: func() *mocks.Parser {
				parser := new(mocks.Parser)
				parser.On("Parse", "INVALID key1").Return(nil, database.ErrInvalidSyntax)
				return parser
			},
			engine:      func() *mocks.Engine { return new(mocks.Engine) },
			expectedRes: "error: parse input failed: invalid syntax",
		},
		{
			name:  "Handle GET Command with Error",
			query: "GET key2",
			parser: func() *mocks.Parser {
				parser := new(mocks.Parser)
				parser.On("Parse", "GET key2").Return(&database.Command{Type: database.CommandGET, Args: []string{"key2"}}, nil)
				return parser
			},
			engine: func() *mocks.Engine {
				engine := new(mocks.Engine)
				engine.On("Get", "key2").Return("", database.ErrKeyNotFound)
				return engine
			},
			expectedRes: "error: key not found",
		},
		{
			name:  "Handle DEL Command with Error",
			query: "DEL key2",
			parser: func() *mocks.Parser {
				parser := new(mocks.Parser)
				parser.On("Parse", "DEL key2").Return(&database.Command{Type: database.CommandDEL, Args: []string{"key2"}}, nil)
				return parser
			},
			engine: func() *mocks.Engine {
				engine := new(mocks.Engine)
				engine.On("Del", "key2").Return(database.ErrKeyNotFound)
				return engine
			},
			expectedRes: "error: key not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := tt.parser()
			engine := tt.engine()

			db := database.New(parser, engine)
			result := db.HandleQuery(tt.query)
			assert.Equal(t, tt.expectedRes, result)

			parser.AssertExpectations(t)
			engine.AssertExpectations(t)
		})
	}
}
