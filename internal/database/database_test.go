package database

import (
	"errors"
	"fmt"
	"testing"

	"github.com/neekrasov/kvdb/internal/database/command"
	"github.com/neekrasov/kvdb/internal/database/storage"
	"github.com/neekrasov/kvdb/internal/database/storage/models"
	dbMock "github.com/neekrasov/kvdb/internal/mocks/database"
	"github.com/neekrasov/kvdb/pkg/config"
	"github.com/neekrasov/kvdb/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDatabase_HandleQuery(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	tests := []struct {
		name         string
		query        string
		parseErr     error
		handlerFunc  func(*models.User, []string) string
		user         *models.User
		expected     string
		prepareMocks func(p *dbMock.Parser, s *dbMock.Storage, us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage, rs *dbMock.RolesStorage)
	}{
		{
			name:     "parse error",
			query:    "invalid query",
			parseErr: errors.New("parse error"),
			expected: "error: parse input failed: parse error",
			prepareMocks: func(p *dbMock.Parser, s *dbMock.Storage, us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage, rs *dbMock.RolesStorage) {
				p.On("Parse", "invalid query").Return(
					&command.Command{Type: command.CommandSET, Args: []string{"key", "value"}},
					errors.New("parse error")).Once()
			},
		},
		{
			name:     "invalid operation",
			query:    "invalid operation",
			expected: "error: parse input failed: invalid operation",
			prepareMocks: func(p *dbMock.Parser, s *dbMock.Storage, us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage, rs *dbMock.RolesStorage) {
				p.On("Parse", "invalid operation").Return(
					nil, errors.New("invalid operation")).Once()
			},
		},
		{
			name:     "admin only command with non-admin user",
			query:    "create_user username password",
			user:     &models.User{Username: "nonadmin"},
			expected: "error: permission denied",
			prepareMocks: func(p *dbMock.Parser, s *dbMock.Storage, us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage, rs *dbMock.RolesStorage) {
				p.On("Parse", "create_user username password").Return(
					&command.Command{Type: command.CommandCREATEUSER, Args: []string{"username", "password"}}, nil).Once()
			},
		},
		{
			name:     "valid command",
			query:    "set key value",
			user:     &models.User{Username: "admin", Cur: models.Role{Set: true}},
			expected: "value",
			prepareMocks: func(p *dbMock.Parser, s *dbMock.Storage, us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage, rs *dbMock.RolesStorage) {
				p.On("Parse", "set key value").Return(
					&command.Command{Type: command.CommandSET, Args: []string{"key", "value"}}, nil).Once()
				s.On("Set", mock.Anything, mock.Anything).Return()
			},
		},
		{
			name:     "successful del command",
			query:    "del key",
			user:     &models.User{Username: "admin", Cur: models.Role{Del: true, Namespace: "default"}},
			expected: "key",
			prepareMocks: func(p *dbMock.Parser, s *dbMock.Storage, us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage, rs *dbMock.RolesStorage) {
				p.On("Parse", "del key").Return(&command.Command{Type: command.CommandDEL, Args: []string{"key"}}, nil).Once()
				s.On("Del", "default:key").Return(nil).Once()
			},
		},
		{
			name:     "successful get command",
			query:    "get key",
			user:     &models.User{Username: "admin", Cur: models.Role{Get: true, Namespace: "default"}},
			expected: "value",
			prepareMocks: func(p *dbMock.Parser, s *dbMock.Storage, us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage, rs *dbMock.RolesStorage) {
				p.On("Parse", "get key").Return(&command.Command{Type: command.CommandGET, Args: []string{"key"}}, nil).Once()
				s.On("Get", "default:key").Return("value", nil).Once()
			},
		},
		{
			name:     "successful set command",
			query:    "set key value",
			user:     &models.User{Username: "admin", Cur: models.Role{Set: true, Namespace: "default"}},
			expected: "value",
			prepareMocks: func(p *dbMock.Parser, s *dbMock.Storage, us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage, rs *dbMock.RolesStorage) {
				p.On("Parse", "set key value").Return(&command.Command{Type: command.CommandSET, Args: []string{"key", "value"}}, nil).Once()
				s.On("Set", "default:key", "value").Return().Once()
			},
		},
		{
			name:     "successful createUser command",
			query:    "create_user username password",
			user:     &models.User{Username: "admin"},
			expected: "admin",
			prepareMocks: func(p *dbMock.Parser, s *dbMock.Storage, us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage, rs *dbMock.RolesStorage) {
				p.On("Parse", "create_user username password").Return(&command.Command{Type: command.CommandCREATEUSER, Args: []string{"username", "password"}}, nil).Once()
				us.On("Create", "username", "password").Return(&models.User{Username: "admin"}, nil).Once()
				us.On("Append", "admin").Return([]string{}, nil).Once()
			},
		},
		{
			name:     "successful assignRole command",
			query:    "assign_role username role",
			user:     &models.User{Username: "admin"},
			expected: "OK",
			prepareMocks: func(p *dbMock.Parser, s *dbMock.Storage, us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage, rs *dbMock.RolesStorage) {
				p.On("Parse", "assign_role username role").Return(&command.Command{Type: command.CommandASSIGNROLE, Args: []string{"username", "role"}}, nil).Once()
				rs.On("Get", "role").Return(&models.Role{Name: "role"}, nil).Once()
				us.On("AssignRole", "username", "role").Return(nil).Once()
			},
		},
		{
			name:     "successful users command",
			query:    "users",
			user:     &models.User{Username: "admin"},
			expected: "user1 user2",
			prepareMocks: func(p *dbMock.Parser, s *dbMock.Storage, us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage, rs *dbMock.RolesStorage) {
				p.On("Parse", "users").Return(&command.Command{Type: command.CommandUSERS, Args: []string{}}, nil).Once()
				us.On("ListUsernames").Return([]string{"user1", "user2"}, nil).Once()
			},
		},
		{
			name:     "successful createRole command",
			query:    "create_role role rwd namespace",
			user:     &models.User{Username: "admin"},
			expected: "OK",
			prepareMocks: func(p *dbMock.Parser, s *dbMock.Storage, us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage, rs *dbMock.RolesStorage) {
				p.On("Parse", "create_role role rwd namespace").Return(&command.Command{Type: command.CommandCREATEROLE, Args: []string{"role", "rwd", "namespace"}}, nil).Once()
				ns.On("Exists", "namespace").Return(true).Once()
				rs.On("Get", "role").Return(nil, models.ErrRoleNotFound).Once()
				rs.On("Save", mock.Anything).Return(nil).Once()
				rs.On("Append", "role").Return(nil, nil).Once()
			},
		},
		{
			name:     "successful Delete command",
			query:    "delete_role role",
			user:     &models.User{Username: "admin"},
			expected: "OK",
			prepareMocks: func(p *dbMock.Parser, s *dbMock.Storage, us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage, rs *dbMock.RolesStorage) {
				p.On("Parse", "delete_role role").Return(&command.Command{Type: command.CommandDELETEROLE, Args: []string{"role"}}, nil).Once()
				us.On("ListUsernames").Return([]string{}, nil).Once()
				rs.On("Delete", "role").Return(nil).Once()
			},
		},
		{
			name:     "successful roles command",
			query:    "roles",
			user:     &models.User{Username: "admin"},
			expected: "role1, role2",
			prepareMocks: func(p *dbMock.Parser, s *dbMock.Storage, us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage, rs *dbMock.RolesStorage) {
				p.On("Parse", "roles").Return(&command.Command{Type: command.CommandROLES, Args: []string{}}, nil).Once()
				rs.On("List").Return([]string{"role1", "role2"}, nil).Once()
			},
		},
		{
			name:     "successful me command",
			query:    "me",
			user:     &models.User{Username: "user", Roles: []string{"role1"}, Cur: models.Role{Namespace: "default", Get: true, Set: true, Del: true}},
			expected: "user: 'user', roles: 'role1', ns: 'default', perms: 'rwd'",
			prepareMocks: func(p *dbMock.Parser, s *dbMock.Storage, us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage, rs *dbMock.RolesStorage) {
				p.On("Parse", "me").Return(&command.Command{Type: command.CommandME, Args: []string{}}, nil).Once()
			},
		},
		{
			name:     "successful createNS command",
			query:    "create_ns namespace",
			user:     &models.User{Username: "admin"},
			expected: "OK",
			prepareMocks: func(p *dbMock.Parser, s *dbMock.Storage, us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage, rs *dbMock.RolesStorage) {
				p.On("Parse", "create_ns namespace").Return(&command.Command{Type: command.CommandCREATENAMESPACE, Args: []string{"namespace"}}, nil).Once()
				ns.On("Save", "namespace").Return(nil).Once()
				ns.On("Append", "namespace").Return(nil, nil).Once()
			},
		},
		{
			name:     "successful deleteNS command",
			query:    "delete_ns namespace",
			user:     &models.User{Username: "admin"},
			expected: "OK",
			prepareMocks: func(p *dbMock.Parser, s *dbMock.Storage, us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage, rs *dbMock.RolesStorage) {
				p.On("Parse", "delete_ns namespace").Return(&command.Command{Type: command.CommandDELETENAMESPACE, Args: []string{"namespace"}}, nil).Once()
				ns.On("Delete", "namespace").Return(nil).Once()
			},
		},
		{
			name:     "successful setNamespace command",
			query:    "set_ns namespace",
			user:     &models.User{Username: "user", Roles: []string{"role1"}, Cur: models.Role{Namespace: "default"}},
			expected: "OK",
			prepareMocks: func(p *dbMock.Parser, s *dbMock.Storage, us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage, rs *dbMock.RolesStorage) {
				p.On("Parse", "set_ns namespace").Return(&command.Command{Type: command.CommandSETNS, Args: []string{"namespace"}}, nil).Once()
				ns.On("Exists", "namespace").Return(true).Once()
				rs.On("Get", "role1").Return(&models.Role{Namespace: "namespace"}, nil).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockParser, mockStorage := dbMock.NewParser(t), dbMock.NewStorage(t)
			mockUserStorage, mockRolesStorage := dbMock.NewUsersStorage(t), dbMock.NewRolesStorage(t)
			mockNamespaceStorage := dbMock.NewNamespacesStorage(t)
			sessionStorage := dbMock.NewSessionStorage(t)

			db := New(
				mockParser,
				mockStorage,
				mockUserStorage,
				mockNamespaceStorage,
				mockRolesStorage,
				sessionStorage,
				config.RootConfig{Username: "admin"},
			)

			tt.prepareMocks(
				mockParser, mockStorage, mockUserStorage,
				mockNamespaceStorage, mockRolesStorage,
			)

			result := db.HandleQuery(tt.user, tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDatabase_Login(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	tests := []struct {
		name         string
		query        string
		user         *models.User
		expected     *models.User
		err          error
		prepareMocks func(p *dbMock.Parser, us *dbMock.UsersStorage, ss *dbMock.SessionStorage)
	}{
		{
			name:  "parse error",
			query: "invalid query",
			err:   fmt.Errorf("parse input failed: %w", errors.New("invalid command: unrecognized command")),
			prepareMocks: func(p *dbMock.Parser, us *dbMock.UsersStorage, ss *dbMock.SessionStorage) {
				p.On("Parse", "invalid query").Return(nil, errors.New("invalid command: unrecognized command"))
			},
		},
		{
			name:  "authentication failed",
			query: "auth username password",
			err:   errors.New("authentication failed"),
			prepareMocks: func(p *dbMock.Parser, us *dbMock.UsersStorage, ss *dbMock.SessionStorage) {
				p.On("Parse", "auth username password").Return(
					&command.Command{
						Type: command.CommandAUTH, Args: []string{"username", "password"},
					}, nil)
				us.On("Authenticate", "username", "password").Return(nil, models.ErrAuthenticationFailed)
			},
		},
		{
			name:  "authentication failed, nil usr",
			query: "auth username password",
			err:   models.ErrAuthenticationRequired,
			prepareMocks: func(p *dbMock.Parser, us *dbMock.UsersStorage, ss *dbMock.SessionStorage) {
				p.On("Parse", "auth username password").Return(
					&command.Command{
						Type: command.CommandAUTH, Args: []string{"username", "password"},
					}, nil)
				us.On("Authenticate", "username", "password").Return(nil, nil)
			},
		},
		{
			name:     "successful login",
			query:    "auth username password",
			user:     &models.User{Username: "username"},
			expected: &models.User{Username: "username", Token: "token"},
			prepareMocks: func(p *dbMock.Parser, us *dbMock.UsersStorage, ss *dbMock.SessionStorage) {
				p.On("Parse", "auth username password").Return(
					&command.Command{
						Type: command.CommandAUTH, Args: []string{"username", "password"},
					}, nil)
				us.On("Authenticate", "username", "password").Return(&models.User{Username: "username"}, nil)
				ss.On("Create", "username").Return("token", nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockParser := dbMock.NewParser(t)
			mockUserStorage := dbMock.NewUsersStorage(t)
			mockSessionStorage := dbMock.NewSessionStorage(t)
			tt.prepareMocks(mockParser, mockUserStorage, mockSessionStorage)

			db := &Database{
				parser:      mockParser,
				userStorage: mockUserStorage,
				sessions:    mockSessionStorage,
			}

			user, err := db.Login(tt.query)
			assert.Equal(t, tt.err, err)
			if tt.expected != nil {
				assert.Equal(t, tt.expected.Username, user.Username)
				assert.NotEmpty(t, user.Token)
			}
		})
	}
}

func TestDatabase_Logout(t *testing.T) {
	mockSessionStorage := storage.NewSessionStorage()
	db := &Database{
		sessions: mockSessionStorage,
	}

	user := &models.User{Username: "username", Token: "token"}
	mockSessionStorage.Create(user.Username)

	result := db.Logout(user, []string{})
	assert.Equal(t, "OK", result)
}
