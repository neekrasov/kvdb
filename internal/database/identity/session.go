package identity

import (
	"errors"
	"sync"
	"time"

	"github.com/neekrasov/kvdb/internal/database/identity/models"
)

var (
	ErrSessionAlreadyExists = errors.New("session already exists")
	ErrExpiresSession       = errors.New("session expired")
)

// SessionStorage - a struct that manages user sessions, including creation, retrieval, and deletion.
type SessionStorage struct {
	mu              sync.RWMutex
	sessions        map[string]models.Session
	sessionLifeTime time.Duration
}

// NewSessionStorage - initializes and returns a new SessionStorage instance.
func NewSessionStorage(defaultExpiration time.Duration) *SessionStorage {
	return &SessionStorage{
		sessions:        make(map[models.SessionID]models.Session),
		sessionLifeTime: defaultExpiration,
	}
}

// Create - creates a new session for a user and stores it in the session storage.
func (s *SessionStorage) Create(id models.SessionID, user *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.sessions[id]; ok {
		return ErrSessionAlreadyExists
	}

	now := time.Now()
	session := models.Session{
		User:      user,
		CreatedAt: now,
	}

	if s.sessionLifeTime > 0 {
		session.ExpiresAt = now.Add(s.sessionLifeTime)
	}

	s.sessions[id] = session
	return nil
}

// Get - retrieves a session by its token.
func (s *SessionStorage) Get(id string) (*models.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[id]
	if !exists {
		return nil, ErrExpiresSession
	}

	if !session.ExpiresAt.IsZero() && time.Now().After(session.ExpiresAt) {
		return nil, ErrExpiresSession
	}

	return &session, nil
}

// Delete - deletes a session by its token.
func (s *SessionStorage) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, id)
}

func (s *SessionStorage) List() []models.Session {
	sessions := make([]models.Session, 0, len(s.sessions))
	for _, s := range s.sessions {
		sessions = append(sessions, s)
	}

	return sessions
}
