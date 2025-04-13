package application

import (
	"github.com/neekrasov/kvdb/internal/database/compute"
)

func initCommandTrie() *compute.TrieNode {
	root := compute.NewTrieNode()
	root.Insert(compute.CommandSET, map[string]compute.CommandParam{
		compute.KeyArg:   {Required: true, Positional: true, Position: 0},
		compute.ValueArg: {Required: true, Positional: true, Position: 1},
		compute.TTLArg:   {Required: false, Positional: false},
		compute.NSArg:    {Required: false, Positional: false},
	})
	root.Insert(compute.CommandGET, map[string]compute.CommandParam{
		compute.KeyArg: {Required: true, Positional: true, Position: 0},
		compute.TTLArg: {Required: false, Positional: false},
		compute.NSArg:  {Required: false, Positional: false},
	})
	root.Insert(compute.CommandDEL, map[string]compute.CommandParam{
		compute.KeyArg: {Required: true, Positional: true, Position: 0},
		compute.TTLArg: {Required: false, Positional: false},
		compute.NSArg:  {Required: false, Positional: false},
	})
	root.Insert(compute.CommandAUTH, map[string]compute.CommandParam{
		compute.UsernameArg: {Required: true, Positional: true, Position: 0},
		compute.PasswordArg: {Required: true, Positional: true, Position: 1},
	})
	root.Insert(compute.CommandCREATEUSER, map[string]compute.CommandParam{
		compute.UsernameArg: {Required: true, Positional: true, Position: 0},
		compute.PasswordArg: {Required: true, Positional: true, Position: 1},
	})
	root.Insert(compute.CommandGETUSER, map[string]compute.CommandParam{
		compute.UsernameArg: {Required: true, Positional: true, Position: 0},
	})
	root.Insert(compute.CommandDELETEUSER, map[string]compute.CommandParam{
		compute.UsernameArg: {Required: true, Positional: true, Position: 0},
	})
	root.Insert(compute.CommandASSIGNROLE, map[string]compute.CommandParam{
		compute.UsernameArg: {Required: true, Positional: true, Position: 0},
		compute.RoleArg:     {Required: true, Positional: true, Position: 1},
	})
	root.Insert(compute.CommandDIVESTROLE, map[string]compute.CommandParam{
		compute.UsernameArg: {Required: true, Positional: true, Position: 0},
		compute.RoleArg:     {Required: true, Positional: true, Position: 1},
	})
	root.Insert(compute.CommandCREATEROLE, map[string]compute.CommandParam{
		compute.RoleNameArg:    {Required: true, Positional: true, Position: 0},
		compute.PermissionsArg: {Required: true, Positional: true, Position: 1},
		compute.NamespaceArg:   {Required: true, Positional: true, Position: 2},
	})
	root.Insert(compute.CommandGETROLE, map[string]compute.CommandParam{
		compute.RoleNameArg: {Required: true, Positional: true, Position: 0},
	})
	root.Insert(compute.CommandDELETEROLE, map[string]compute.CommandParam{
		compute.RoleNameArg: {Required: true, Positional: true, Position: 0},
	})
	root.Insert(compute.CommandCREATENAMESPACE, map[string]compute.CommandParam{
		compute.NamespaceArg: {Required: true, Positional: true, Position: 0},
	})
	root.Insert(compute.CommandDELETENAMESPACE, map[string]compute.CommandParam{
		compute.NamespaceArg: {Required: true, Positional: true, Position: 0},
	})
	root.Insert(compute.CommandSETNS, map[string]compute.CommandParam{
		compute.NamespaceArg: {Required: true, Positional: true, Position: 0},
	})
	root.Insert(compute.CommandUSERS, nil)
	root.Insert(compute.CommandME, nil)
	root.Insert(compute.CommandROLES, nil)
	root.Insert(compute.CommandNAMESPACES, nil)
	root.Insert(compute.CommandSESSIONS, nil)
	root.Insert(compute.CommandHELP, nil)
	root.Insert(compute.CommandWATCH, map[string]compute.CommandParam{
		compute.KeyArg: {Required: true, Positional: true, Position: 0},
		compute.NSArg:  {Required: false, Positional: false},
	})
	root.Insert(compute.CommandSTAT, nil)

	return root
}
