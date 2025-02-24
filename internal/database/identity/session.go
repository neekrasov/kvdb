package identity

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

// Session - struct that represents a user session, including the username, token, and expiration time.
type Session struct {
	Username  string
	Token     string
	ExpiresAt time.Time
}

// GenerateToken - generates a random token for session management.
func GenerateToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// SessionStorage - a struct that manages user sessions, including creation, retrieval, and deletion.
type SessionStorage struct {
	mu       sync.RWMutex
	sessions map[string]Session
}

// NewSessionStorage - initializes and returns a new SessionStorage instance.
func NewSessionStorage() *SessionStorage {
	return &SessionStorage{
		sessions: make(map[string]Session),
	}
}

// Create - creates a new session for a user and stores it in the session storage.
func (s *SessionStorage) Create(username string) (string, error) {
	token, err := GenerateToken()
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[token] = Session{
		Token:     token,
		Username:  username,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	return token, nil
}

// Get - retrieves a session by its token.
func (s *SessionStorage) Get(token string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[token]
	if !exists || time.Now().After(session.ExpiresAt) {
		return nil, errors.New("invalid or expired token")
	}

	return &session, nil
}

// Delete - deletes a session by its token.
func (s *SessionStorage) Delete(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, token)
	return nil
}
