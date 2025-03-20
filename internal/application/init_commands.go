package application

import (
	"github.com/neekrasov/kvdb/internal/database/compute"
)

func initCommandTrie() *compute.TrieNode {
	root := compute.NewTrieNode()
	root.Insert(compute.CommandCREATEROLE, "create", "role")
	root.Insert(compute.CommandCREATEUSER, "create", "user")
	root.Insert(compute.CommandGETROLE, "get", "role")
	root.Insert(compute.CommandGETUSER, "get", "user")
	root.Insert(compute.CommandCREATEUSER, "delete", "user")
	root.Insert(compute.CommandASSIGNROLE, "assign", "role")
	root.Insert(compute.CommandDIVESTROLE, "divest", "role")
	root.Insert(compute.CommandDELETEROLE, "delete", "role")
	root.Insert(compute.CommandCREATENAMESPACE, "create", "ns")
	root.Insert(compute.CommandDELETENAMESPACE, "delete", "ns")
	root.Insert(compute.CommandSETNS, "set", "ns")
	root.Insert(compute.CommandGET, "get")
	root.Insert(compute.CommandSET, "set")
	root.Insert(compute.CommandDEL, "del")
	root.Insert(compute.CommandAUTH, "login")
	root.Insert(compute.CommandUSERS, "users")
	root.Insert(compute.CommandME, "me")
	root.Insert(compute.CommandROLES, "roles")
	root.Insert(compute.CommandNAMESPACES, "ns")
	root.Insert(compute.CommandSESSIONS, "sessions")
	root.Insert(compute.CommandHELP, "help")
	root.Insert(compute.CommandWATCH, "watch")

	return root
}
