package application

import (
	"github.com/neekrasov/kvdb/internal/database/compute"
)

func initCommandTrie() *compute.TrieNode {
	root := compute.NewTrieNode()
	root.Insert([]string{"create", "role"}, compute.CommandCREATEROLE)
	root.Insert([]string{"create", "user"}, compute.CommandCREATEUSER)
	root.Insert([]string{"assign", "role"}, compute.CommandASSIGNROLE)
	root.Insert([]string{"delete", "role"}, compute.CommandDELETEROLE)
	root.Insert([]string{"create", "ns"}, compute.CommandCREATENAMESPACE)
	root.Insert([]string{"delete", "ns"}, compute.CommandDELETENAMESPACE)
	root.Insert([]string{"set", "ns"}, compute.CommandSETNS)
	root.Insert([]string{"get"}, compute.CommandGET)
	root.Insert([]string{"set"}, compute.CommandSET)
	root.Insert([]string{"del"}, compute.CommandDEL)
	root.Insert([]string{"login"}, compute.CommandAUTH)
	root.Insert([]string{"users"}, compute.CommandUSERS)
	root.Insert([]string{"me"}, compute.CommandME)
	root.Insert([]string{"roles"}, compute.CommandROLES)
	root.Insert([]string{"ns"}, compute.CommandNAMESPACES)
	root.Insert([]string{"help"}, compute.CommandHELP)

	return root
}
