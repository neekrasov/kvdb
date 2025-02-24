package models

import (
	"errors"
	"fmt"
	"strings"
)

const (
	RootRoleName    = "root"
	DefaultRoleName = "default"
)

var DefaultRole = Role{
	Name: DefaultRoleName,
	Get:  true, Set: true, Del: true,
	Namespace: DefaultNameSpace,
}

var ErrInvalidPerms = errors.New("invalid perms: perms must contain only 'r', 'w', 'd'")

// Role - struct representing a role in the system.
type Role struct {
	Name      string `json:"name"`
	Get       bool   `json:"get"`
	Set       bool   `json:"set"`
	Del       bool   `json:"del"`
	Namespace string `json:"namespace"`
}

// Perms - returns a string representation of the role's permissions (r for read, w for write, d for delete).
func (r *Role) Perms() string {
	var res string

	if r.Get {
		res += "r"
	}

	if r.Set {
		res += "w"
	}

	if r.Del {
		res += "d"
	}

	return res
}

// String - returns a formatted string representation of the role, including its name, permissions, and namespace.
func (r *Role) String() string {
	return fmt.Sprintf(
		"name: '%s', perms: '%s' namespace: '%s'",
		r.Name, r.Perms(), r.Namespace,
	)
}

// NewRole - creates a new role with the specified name, permissions, and namespace.
func NewRole(name, perms, namespace string) (Role, error) {
	if len(perms) > 3 || len(perms) == 0 {
		return Role{}, errors.New("perms must be between 1 and 3 characters")
	}

	for _, char := range perms {
		if char != 'r' && char != 'w' && char != 'd' {
			return Role{}, ErrInvalidPerms
		}
	}

	if namespace == "" {
		namespace = DefaultNameSpace
	}

	return Role{
		Name:      name,
		Namespace: namespace,
		Get:       strings.ContainsRune(perms, 'r'),
		Set:       strings.ContainsRune(perms, 'w'),
		Del:       strings.ContainsRune(perms, 'd'),
	}, nil
}
