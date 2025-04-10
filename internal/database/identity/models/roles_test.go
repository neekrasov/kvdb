package models_test

import (
	"testing"

	"github.com/neekrasov/kvdb/internal/config"
	"github.com/neekrasov/kvdb/internal/database/identity/models"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestModels(t *testing.T) {
	t.Parallel()

	t.Run("NewRole - valid perms", func(t *testing.T) {
		role, err := models.NewRole("admin", "rwd", "")
		assert.NoError(t, err)
		assert.Equal(t, "admin", role.Name)
		assert.Equal(t, "default", role.Namespace)
		assert.Equal(t, role.String(), "name: 'admin', perms: 'rwd' namespace: 'default'")
		assert.True(t, role.Get)
		assert.True(t, role.Set)
	})
	t.Run("NewRole - invalid perms", func(t *testing.T) {
		_, err := models.NewRole("admin", "xyz", "default")
		assert.Error(t, err)
	})
	t.Run("NewRole - invalid perms len", func(t *testing.T) {
		_, err := models.NewRole("admin", "rwwww", "default")
		assert.Error(t, err)
	})
	t.Run("Role Perms", func(t *testing.T) {
		role := models.Role{Name: "user", Get: true, Set: true}
		assert.Equal(t, "rw", role.Perms())

	})
	t.Run("IsAdmin", func(t *testing.T) {
		hashedPassword, err := bcrypt.GenerateFromPassword(
			[]byte("password"), bcrypt.DefaultCost)
		assert.NoError(t, err)

		cfg := &config.RootConfig{Username: "admin", Password: "password"}
		user := models.User{Username: "admin", Password: string(hashedPassword)}
		assert.True(t, user.IsAdmin(cfg))
	})
}
