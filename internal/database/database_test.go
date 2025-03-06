package database

import (
	"errors"
	"fmt"
	"testing"

	"github.com/neekrasov/kvdb/internal/config"
	"github.com/neekrasov/kvdb/internal/database/compute"
	"github.com/neekrasov/kvdb/internal/database/identity"
	"github.com/neekrasov/kvdb/internal/database/identity/models"
	dbMock "github.com/neekrasov/kvdb/internal/mocks/database"
	"github.com/neekrasov/kvdb/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDatabase_HandleQuery(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	sessionID := "test"
	session := &models.Session{
		User: &models.User{
			Username: "admin",
			ActiveRole: models.Role{
				Get:       true,
				Set:       true,
				Del:       true,
				Namespace: models.DefaultNameSpace,
			},
		},
	}

	tests := []struct {
		name         string
		query        string
		parseErr     error
		expected     string
		prepareMocks func(
			p *dbMock.Parser, s *dbMock.Storage, us *dbMock.UsersStorage,
			ns *dbMock.NamespacesStorage, rs *dbMock.RolesStorage,
			ss *dbMock.SessionStorage,
		)
	}{
		{
			name:     "parse error",
			query:    "invalid query",
			parseErr: errors.New("parse error"),
			expected: fmt.Sprintf("%s parse input failed: parse error", errPrefix),
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage, us *dbMock.UsersStorage,
				ns *dbMock.NamespacesStorage, rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(session, nil).Once()
				p.On("Parse", "invalid query").Return(
					&compute.Command{Type: compute.CommandSET, Args: []string{"key", "value"}},
					errors.New("parse error")).Once()
			},
		},
		{
			name:     "invalid operation",
			query:    "invalid operation",
			expected: fmt.Sprintf("%s parse input failed: invalid operation", errPrefix),
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage, us *dbMock.UsersStorage,
				ns *dbMock.NamespacesStorage, rs *dbMock.RolesStorage,
				ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(session, nil).Once()
				p.On("Parse", "invalid operation").Return(
					nil, errors.New("invalid operation")).Once()
			},
		},
		{
			name:     "admin only command with non-admin user",
			query:    compute.CommandCREATEUSER.String() + " username password",
			expected: fmt.Sprintf("%s permission denied", errPrefix),
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(&models.Session{
					User: &models.User{Username: "non-admin"},
				}, nil).Once()
				p.On("Parse", compute.CommandCREATEUSER.String()+" username password").Return(
					&compute.Command{Type: compute.CommandCREATEUSER,
						Args: []string{"username", "password"}}, nil).Once()
			},
		},
		{
			name:     "valid command",
			query:    compute.CommandSET.Make("key", "value"),
			expected: okPrefix,
			prepareMocks: func(p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(session, nil).Once()
				p.On("Parse", compute.CommandSET.Make("key", "value")).Return(
					&compute.Command{Type: compute.CommandSET,
						Args: []string{"key", "value"}}, nil).Once()
				s.On("Set", mock.Anything, mock.Anything).Return(nil).Once()
			},
		},
		{
			name:     "successful del command",
			query:    compute.CommandDEL.Make("key"),
			expected: okPrefix,
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(session, nil).Once()
				p.On("Parse", compute.CommandDEL.Make("key")).Return(
					&compute.Command{Type: compute.CommandDEL,
						Args: []string{"key"}}, nil).Once()
				s.On("Del", "default:key").Return(nil).Once()
			},
		},
		{
			name:     "successful get command",
			query:    compute.CommandGET.Make("key"),
			expected: okPrefix + " value",
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(session, nil).Once()
				p.On("Parse", compute.CommandGET.Make("key")).Return(
					&compute.Command{Type: compute.CommandGET,
						Args: []string{"key"}}, nil).Once()
				s.On("Get", "default:key").Return("value", nil).Once()
			},
		},
		{
			name:     "successful set command",
			query:    compute.CommandSET.Make("key", "value"),
			expected: okPrefix,
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(session, nil).Once()
				p.On("Parse", compute.CommandSET.String()+" key value").Return(
					&compute.Command{Type: compute.CommandSET,
						Args: []string{"key", "value"}}, nil).Once()
				s.On("Set", "default:key", "value").Return(nil).Once()
			},
		},
		{
			name:     "successful create user command",
			query:    compute.CommandCREATEUSER.Make("username", "password"),
			expected: okPrefix,
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(session, nil).Once()
				p.On("Parse", compute.CommandCREATEUSER.Make("username", "password")).Return(
					&compute.Command{Type: compute.CommandCREATEUSER,
						Args: []string{"username", "password"}}, nil).Once()
				us.On("Create", "username", "password").Return(&models.User{Username: "admin"}, nil).Once()
				us.On("Append", "admin").Return([]string{}, nil).Once()
			},
		},
		{
			name:     "successful assign role command",
			query:    compute.CommandASSIGNROLE.Make("username", "role"),
			expected: okPrefix,
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(session, nil).Once()
				p.On("Parse", compute.CommandASSIGNROLE.Make("username", "role")).Return(
					&compute.Command{Type: compute.CommandASSIGNROLE,
						Args: []string{"username", "role"}}, nil).Once()
				us.On("AssignRole", "username", "role").Return(nil).Once()
			},
		},
		{
			name:     "successful users command",
			query:    compute.CommandUSERS.String(),
			expected: fmt.Sprintf("%s [\"user1\",\"user2\"]", okPrefix),
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(session, nil).Once()
				p.On("Parse", compute.CommandUSERS.String()).Return(
					&compute.Command{Type: compute.CommandUSERS,
						Args: []string{}}, nil).Once()
				us.On("ListUsernames").Return([]string{"user1", "user2"}, nil).Once()
			},
		},
		{
			name:     "successful create role command",
			query:    compute.CommandCREATEROLE.Make("role", "rwd", "namespace"),
			expected: okPrefix,
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(session, nil).Once()
				p.On("Parse", compute.CommandCREATEROLE.String()+" role rwd namespace").Return(
					&compute.Command{Type: compute.CommandCREATEROLE,
						Args: []string{"role", "rwd", "namespace"}}, nil).Once()
				ns.On("Exists", "namespace").Return(true).Once()
				rs.On("Get", "role").Return(nil, identity.ErrRoleNotFound).Once()
				rs.On("Save", mock.Anything).Return(nil).Once()
				rs.On("Append", "role").Return(nil, nil).Once()
			},
		},
		{
			name:     "successful delete role command",
			query:    compute.CommandDELETEROLE.Make("role"),
			expected: okPrefix,
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(session, nil).Once()
				p.On("Parse", compute.CommandDELETEROLE.String()+" role").Return(
					&compute.Command{Type: compute.CommandDELETEROLE,
						Args: []string{"role"}}, nil).Once()
				us.On("ListUsernames").Return([]string{}, nil).Once()
				rs.On("Delete", "role").Return(nil).Once()
			},
		},
		{
			name:     "successful roles command",
			query:    compute.CommandROLES.String(),
			expected: fmt.Sprintf("%s [\"role1\",\"role2\"]", okPrefix),
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(session, nil).Once()
				p.On("Parse", compute.CommandROLES.String()).Return(
					&compute.Command{Type: compute.CommandROLES,
						Args: []string{}}, nil).Once()
				rs.On("List").Return([]string{"role1", "role2"}, nil).Once()
			},
		},
		{
			name:     "successful me command",
			query:    compute.CommandME.String(),
			expected: fmt.Sprintf("%s user: 'user', roles: 'role1', ns: 'default', perms: 'rwd'", okPrefix),
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(
					&models.Session{
						User: &models.User{
							Username: "user",
							Roles:    []string{"role1"},
							ActiveRole: models.Role{
								Namespace: "default",
								Get:       true, Set: true, Del: true}},
					}, nil).Once()
				p.On("Parse", compute.CommandME.String()).Return(
					&compute.Command{Type: compute.CommandME,
						Args: []string{}}, nil).Once()
			},
		},
		{
			name:     "successful create ns command",
			query:    compute.CommandCREATENAMESPACE.Make("namespace"),
			expected: okPrefix,
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(session, nil).Once()
				p.On("Parse", compute.CommandCREATENAMESPACE.Make("namespace")).Return(
					&compute.Command{Type: compute.CommandCREATENAMESPACE,
						Args: []string{"namespace"}}, nil).Once()
				ns.On("Save", "namespace").Return(nil).Once()
				ns.On("Append", "namespace").Return(nil, nil).Once()
			},
		},
		{
			name:     "successful delete ns command",
			query:    compute.CommandDELETENAMESPACE.Make("namespace"),
			expected: okPrefix,
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(session, nil).Once()
				rs.On("List").Return(nil, nil)
				p.On("Parse", compute.CommandDELETENAMESPACE.Make("namespace")).Return(
					&compute.Command{Type: compute.CommandDELETENAMESPACE,
						Args: []string{"namespace"}}, nil).Once()
				ns.On("Delete", "namespace").Return(nil).Once()
			},
		},
		{
			name:     "successful set namespace command",
			query:    compute.CommandSETNS.Make("namespace"),
			expected: okPrefix,
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(&models.Session{
					User: &models.User{
						Username: "user", Roles: []string{"role1"},
						ActiveRole: models.Role{Namespace: "default"}},
				}, nil).Once()
				p.On("Parse", compute.CommandSETNS.Make("namespace")).Return(
					&compute.Command{Type: compute.CommandSETNS,
						Args: []string{"namespace"}}, nil).Once()
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
			mockSessionStorage := dbMock.NewSessionStorage(t)

			tt.prepareMocks(
				mockParser, mockStorage, mockUserStorage,
				mockNamespaceStorage, mockRolesStorage,
				mockSessionStorage,
			)

			db := New(
				mockParser, mockStorage, mockUserStorage,
				mockNamespaceStorage, mockRolesStorage,
				mockSessionStorage, &config.RootConfig{Username: "admin"},
			)

			result := db.HandleQuery(sessionID, tt.query)
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
			prepareMocks: func(
				p *dbMock.Parser, us *dbMock.UsersStorage,
				ss *dbMock.SessionStorage,
			) {
				p.On("Parse", "invalid query").Return(nil, errors.New("invalid command: unrecognized command"))
			},
		},
		{
			name:  "authentication failed",
			query: compute.CommandAUTH.String() + " username password",
			err:   errors.New("authentication failed"),
			prepareMocks: func(
				p *dbMock.Parser, us *dbMock.UsersStorage,
				ss *dbMock.SessionStorage,
			) {
				p.On("Parse", compute.CommandAUTH.String()+" username password").Return(
					&compute.Command{
						Type: compute.CommandAUTH,
						Args: []string{"username", "password"},
					}, nil)
				us.On("Authenticate", "username", "password").Return(nil, identity.ErrAuthenticationFailed)
			},
		},
		{
			name:  "authentication failed, nil usr",
			query: compute.CommandAUTH.String() + " username password",
			err:   ErrAuthenticationRequired,
			prepareMocks: func(
				p *dbMock.Parser, us *dbMock.UsersStorage,
				ss *dbMock.SessionStorage,
			) {
				p.On("Parse", compute.CommandAUTH.String()+" username password").Return(
					&compute.Command{
						Type: compute.CommandAUTH,
						Args: []string{"username", "password"},
					}, nil)
				us.On("Authenticate", "username", "password").Return(nil, nil)
			},
		},
		{
			name:     "successful login",
			query:    compute.CommandAUTH.String() + " username password",
			user:     &models.User{Username: "username"},
			expected: &models.User{Username: "username"},
			prepareMocks: func(
				p *dbMock.Parser, us *dbMock.UsersStorage,
				ss *dbMock.SessionStorage,
			) {
				user := &models.User{Username: "username"}
				p.On("Parse", compute.CommandAUTH.String()+" username password").Return(
					&compute.Command{
						Type: compute.CommandAUTH, Args: []string{"username", "password"},
					}, nil)
				us.On("Authenticate", "username", "password").Return(user, nil)
				ss.On("Create", "1", user).Return(nil)
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

			user, err := db.Login("1", tt.query)
			assert.Equal(t, tt.err, err)
			if tt.expected != nil {
				assert.Equal(t, tt.expected.Username, user.Username)

			}
		})
	}
}

func TestDatabase_Logout(t *testing.T) {
	mockSessionStorage := identity.NewSessionStorage()
	db := &Database{sessions: mockSessionStorage}

	user := &models.User{Username: "username"}
	err := mockSessionStorage.Create("1", user)
	require.NoError(t, err)

	result := db.Logout("1")
	assert.Equal(t, okPrefix, result)
}
