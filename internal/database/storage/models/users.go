package models

import "github.com/neekrasov/kvdb/pkg/config"

// User - A struct representing a user in the system.
type User struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Roles    []string `json:"role"`
	Cur      Role     `json:"cur"`
	Token    string   `json:"-"`
}

// IsAdmin - Checks if the user is an admin by comparing their username and password with the system's root configuration.
func (u *User) IsAdmin(cfg *config.RootConfig) bool {
	return u.Username == cfg.Username && u.Password == cfg.Password
}
