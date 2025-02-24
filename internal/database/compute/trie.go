package compute

type TrieNode struct {
	children map[string]*TrieNode
	command  CommandType
}

func NewTrieNode() *TrieNode {
	return &TrieNode{
		children: make(map[string]*TrieNode),
	}
}

func (t *TrieNode) Insert(command []string, cmdType CommandType) {
	current := t
	for _, part := range command {
		if _, exists := current.children[part]; !exists {
			current.children[part] = NewTrieNode()
		}
		current = current.children[part]
	}
	current.command = cmdType
}

func (t *TrieNode) Search(tokens []string) (CommandType, []string) {
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
