package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSessionStorage(t *testing.T) {
	sessStorage := NewSessionStorage()
	username := "testUser"
	t.Run("Create session", func(t *testing.T) {
		token, err := sessStorage.Create(username)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})
	t.Run("Get session - valid", func(t *testing.T) {
		token, _ := sessStorage.Create(username)
		sess, err := sessStorage.Get(token)
		assert.NoError(t, err)
		assert.Equal(t, username, sess.Username)
	})
	t.Run("Get session - expired", func(t *testing.T) {
		token, _ := sessStorage.Create(username)
		sessStorage.sessions[token] = Session{Username: username, Token: token, ExpiresAt: time.Now().Add(-1 * time.Hour)}
		_, err := sessStorage.Get(token)
		assert.Error(t, err)
	})
	t.Run("Delete session", func(t *testing.T) {
		token, _ := sessStorage.Create(username)
		err := sessStorage.Delete(token)
		assert.NoError(t, err)
	})

	t.Run("Generate token", func(t *testing.T) {
		token, err := GenerateToken()
		assert.NoError(t, err)
		assert.Len(t, token, 32)
	})
}
