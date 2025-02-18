package compute

import (
	"fmt"
	"strings"

	database "github.com/neekrasov/kvdb/internal/database"
	"github.com/neekrasov/kvdb/internal/database/models"
	"github.com/neekrasov/kvdb/pkg/logger"
	"go.uber.org/zap"
)

// Parser - parses queries into commands.
type Parser struct {
	trie *TrieNode
}

// NewParser creates and returns a new instance of Parser.
func NewParser(trie *TrieNode) *Parser {
	return &Parser{
		trie: trie,
	}
}

// Parse converts the query string into a Command or returns an error for invalid syntax.
func (p *Parser) Parse(query string) (*models.Command, error) {
	if query == "" {
		return nil, fmt.Errorf("%w: query cannot be empty", database.ErrInvalidSyntax)
	}

	tokens := strings.Fields(query)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("%w: query cannot be empty", database.ErrInvalidSyntax)
	}

	logger.Debug("parsed tokens", zap.Strings("tokens", tokens))
	commandType, args := p.trie.Search(tokens)
	return models.NewCommand(commandType, args)
}
