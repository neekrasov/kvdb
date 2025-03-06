package models

import (
	"math/rand"
	"time"
)

var charset = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

type SessionID = string

// Session - struct that represents a user session, including the username, and expiration time.
type Session struct {
	User      *User     `json:"user"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func GenSessionID(length int) SessionID {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return SessionID(b)
}
