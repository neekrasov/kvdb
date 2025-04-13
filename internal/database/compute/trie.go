package compute

import (
	"fmt"
	"sort"
	"strings"
)

type TrieNode struct {
	children map[string]*TrieNode
	command  CommandType
	params   map[string]CommandParam
}

func NewTrieNode() *TrieNode {
	return &TrieNode{
		children: make(map[string]*TrieNode),
		params:   make(map[string]CommandParam),
	}
}

func (t *TrieNode) Insert(cmdType CommandType, params map[string]CommandParam) {
	current := t
	for _, part := range cmdType.Split() {
		if _, exists := current.children[part]; !exists {
			current.children[part] = NewTrieNode()
		}
		current = current.children[part]
	}
	current.command = cmdType
	current.params = params
}

func (t *TrieNode) Search(tokens []string) (CommandType, map[string]string, error) {
	current := t
	consumedTokens := 0

	// Traverse the trie with tokens
	for _, token := range tokens {
		if next, exists := current.children[token]; exists {
			current = next
			consumedTokens++
		} else {
			break
		}
	}

	if current.command == "" {
		return "", nil, fmt.Errorf("%w: unknown command", ErrInvalidCommand)
	}

	args := make(map[string]string)
	remainingTokens := tokens[consumedTokens:]

	// Process positional parameters first
	if len(current.params) > 0 {
		positionalParams := make([]struct {
			name  string
			param CommandParam
		}, 0, len(current.params))

		for name, param := range current.params {
			if param.Positional {
				positionalParams = append(positionalParams, struct {
					name  string
					param CommandParam
				}{name, param})
			}
		}

		sort.Slice(positionalParams, func(i, j int) bool {
			return positionalParams[i].param.Position < positionalParams[j].param.Position
		})

		for _, pp := range positionalParams {
			if len(remainingTokens) == 0 {
				break
			}
			args[pp.name] = remainingTokens[0]
			remainingTokens = remainingTokens[1:]
		}
	}

	// Process named parameters
	for i := 0; i < len(remainingTokens); {
		if i+1 >= len(remainingTokens) {
			return "", nil, fmt.Errorf("%w: missing value for parameter '%s'",
				ErrInvalidSyntax, remainingTokens[i])
		}

		paramName := strings.ToLower(remainingTokens[i])
		param, exists := current.params[paramName]
		if !exists || param.Positional {
			return "", nil, fmt.Errorf("%w: unknown parameter '%s'",
				ErrInvalidSyntax, remainingTokens[i])
		}

		args[paramName] = remainingTokens[i+1]
		i += 2
	}

	// Validate required parameters
	for name, param := range current.params {
		if param.Required {
			if _, exists := args[name]; !exists {
				return "", nil, fmt.Errorf("%w: missing required parameter '%s'",
					ErrInvalidSyntax, name)
			}
		}
	}

	return current.command, args, nil
}
