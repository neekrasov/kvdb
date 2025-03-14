package identity

import (
	"testing"
	"time"

	"github.com/neekrasov/kvdb/internal/database/identity/models"
	"github.com/stretchr/testify/assert"
)

func TestSessionStorage(t *testing.T) {
	t.Parallel()

	sessStorage := NewSessionStorage()
	t.Run("Create session", func(t *testing.T) {
		err := sessStorage.Create("1", &models.User{})
		assert.NoError(t, err)
	})

	t.Run("Create session already exists", func(t *testing.T) {
		err := sessStorage.Create("1", &models.User{})
		assert.ErrorIs(t, err, ErrSessionAlreadyExists)
	})

	t.Run("Get session - valid", func(t *testing.T) {
		username := "testUser2"
		id := "2"
		err := sessStorage.Create(id,
			&models.User{Username: username})
		assert.NoError(t, err)

		sess, err := sessStorage.Get(id)
		assert.NoError(t, err)
		assert.Equal(t, username, sess.User.Username)
	})

	t.Run("Get session - Expired", func(t *testing.T) {
		username := "testUser2"
		sessStorage.sessions[username] = models.Session{
			User:      &models.User{Username: username},
			ExpiresAt: time.Now().Add(-time.Hour),
		}

		sess, err := sessStorage.Get(username)
		assert.ErrorIs(t, err, ErrExpiresSession)

		assert.Nil(t, sess)
	})

	t.Run("Delete session", func(t *testing.T) {
		id := "3"
		err := sessStorage.Create(id, &models.User{})
		assert.NoError(t, err)

		sessStorage.Delete(id)

		_, err = sessStorage.Get(id)
		assert.ErrorIs(t, err, ErrExpiresSession)
	})

	t.Run("List sessions", func(t *testing.T) {
		id := "4"
		err := sessStorage.Create(id, &models.User{})
		assert.NoError(t, err)

		sessions := sessStorage.List()
		assert.NotEmpty(t, sessions)
	})
}
