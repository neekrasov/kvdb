package compute

import (
	"github.com/neekrasov/kvdb/internal/database/models"
)

type TrieNode struct {
	children map[string]*TrieNode
	command  models.CommandType
}

func NewTrieNode() *TrieNode {
	return &TrieNode{
		children: make(map[string]*TrieNode),
	}
}

func (t *TrieNode) Insert(command []string, cmdType models.CommandType) {
	current := t
	for _, part := range command {
		if _, exists := current.children[part]; !exists {
			current.children[part] = NewTrieNode()
		}
		current = current.children[part]
	}
	current.command = cmdType
}

func (t *TrieNode) Search(tokens []string) (models.CommandType, []string) {
	current := t
	consumedTokens := 0

	for _, token := range tokens {
		if next, exists := current.children[token]; exists {
			current = next
			consumedTokens++
		} else {
			break
		}
	}

	return current.command, tokens[consumedTokens:]
}
