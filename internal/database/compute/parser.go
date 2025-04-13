package compute

import (
	"fmt"
	"strings"

	"github.com/neekrasov/kvdb/pkg/logger"
	"go.uber.org/zap"
)

// Parser - parses queries into commands.
type Parser struct {
	trie *TrieNode
}

// NewParser - creates and returns a new instance of Parser.
func NewParser(trie *TrieNode) *Parser {
	return &Parser{trie: trie}
}

// Parse - converts the query string into a Command or returns an error for invalid syntax.
func (p *Parser) Parse(query string) (*Command, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("%w: query cannot be empty", ErrInvalidSyntax)
	}

	tokens := strings.Fields(query)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("%w: query cannot be empty", ErrInvalidSyntax)
	}

	commandType, args, err := p.trie.Search(tokens)
	if err != nil {
		return nil, err
	}

	logger.Debug("command successfully created",
		zap.Stringer("cmd_type", commandType),
		zap.Any("args", args))

	return NewCommand(commandType, args)
}
