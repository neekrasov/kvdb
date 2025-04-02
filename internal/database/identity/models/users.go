package models

import (
	"github.com/neekrasov/kvdb/internal/config"
	"golang.org/x/crypto/bcrypt"
)

// User - struct representing a user in the system.
type User struct {
	Username   string   `json:"username"`
	Password   string   `json:"-"`
	Roles      []string `json:"roles"`
	ActiveRole Role     `json:"role"`
}

// IsAdmin - checks if the user is an admin by comparing their username and password with the system's root configuration.
func (u *User) IsAdmin(cfg *config.RootConfig) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(cfg.Password))
	if err != nil {
		return false
	}

	return u.Username == cfg.Username
}
