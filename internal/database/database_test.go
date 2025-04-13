package database

import (
	"context"
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
	adminSession := &models.Session{
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
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				p.On("Parse", "invalid query").Return(nil, errors.New("parse error")).Once()
			},
		},
		{
			name:     "invalid operation (nil command)",
			query:    "invalid operation",
			expected: fmt.Sprintf("%s parse input failed: invalid operation", errPrefix),
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage, us *dbMock.UsersStorage,
				ns *dbMock.NamespacesStorage, rs *dbMock.RolesStorage,
				ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				p.On("Parse", "invalid operation").Return(
					nil, errors.New("invalid operation")).Once()
			},
		},
		{
			name:     "admin only command with non-admin user",
			query:    compute.CommandCREATEUSER.Make("username", "password"),
			expected: fmt.Sprintf("%s permission denied", errPrefix),
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(&models.Session{
					User: &models.User{Username: "non-admin"},
				}, nil).Once()
				p.On("Parse", compute.CommandCREATEUSER.Make("username", "password")).Return(
					&compute.Command{
						Type: compute.CommandCREATEUSER,
						Args: map[string]string{
							compute.UsernameArg: "username",
							compute.PasswordArg: "password",
						},
					}, nil).Once()
			},
		},
		{
			name:     "valid command (SET with defaults)",
			query:    compute.CommandSET.Make("key", "value"),
			expected: okPrefix,
			prepareMocks: func(p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				p.On("Parse", compute.CommandSET.Make("key", "value")).Return(
					&compute.Command{
						Type: compute.CommandSET,
						Args: map[string]string{
							compute.KeyArg:   "key",
							compute.ValueArg: "value",
						},
					}, nil).Once()
				s.On("Set", mock.Anything, "default:key", "value").Return(nil).Once()
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
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				p.On("Parse", compute.CommandDEL.Make("key")).Return(
					&compute.Command{
						Type: compute.CommandDEL,
						Args: map[string]string{
							compute.KeyArg: "key",
						},
					}, nil).Once()
				s.On("Del", mock.Anything, "default:key").Return(nil).Once()
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
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				p.On("Parse", compute.CommandGET.Make("key")).Return(
					&compute.Command{
						Type: compute.CommandGET,
						Args: map[string]string{
							compute.KeyArg: "key",
						},
					}, nil).Once()
				s.On("Get", mock.Anything, "default:key").Return("value", nil).Once()
			},
		},
		{
			name:     "successful set command with TTL and NS args",
			query:    compute.CommandSET.Make("key", "value") + " TTL 10s NS otherns",
			expected: okPrefix,
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				sessionWithOtherNsPerm := &models.Session{
					User: &models.User{
						Username: "admin",
						ActiveRole: models.Role{
							Get: true, Set: true, Del: true,
							Namespace: "otherns",
						},
					},
				}
				ss.On("Get", sessionID).Return(sessionWithOtherNsPerm, nil).Once()
				p.On("Parse", compute.CommandSET.Make("key", "value")+" TTL 10s NS otherns").Return(
					&compute.Command{
						Type: compute.CommandSET,
						Args: map[string]string{
							compute.KeyArg:   "key",
							compute.ValueArg: "value",
							compute.TTLArg:   "10s",
							compute.NSArg:    "otherns",
						},
					}, nil).Once()
				ns.On("Exists", mock.Anything, "otherns").Return(true)
				s.On("Set", mock.Anything, "otherns:key", "value").Return(nil).Once()
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
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				p.On("Parse", compute.CommandCREATEUSER.Make("username", "password")).Return(
					&compute.Command{
						Type: compute.CommandCREATEUSER,
						Args: map[string]string{
							compute.UsernameArg: "username",
							compute.PasswordArg: "password",
						},
					}, nil).Once()
				us.On("Create", mock.Anything, "username", "password").Return(&models.User{Username: "username"}, nil).Once()
				us.On("Append", mock.Anything, "username").Return(nil, nil).Once()
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
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				p.On("Parse", compute.CommandASSIGNROLE.Make("username", "role")).Return(
					&compute.Command{
						Type: compute.CommandASSIGNROLE,
						Args: map[string]string{
							compute.UsernameArg: "username",
							compute.RoleArg:     "role",
						},
					}, nil).Once()
				us.On("AssignRole", mock.Anything, "username", "role").Return(nil).Once()
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
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				p.On("Parse", compute.CommandUSERS.String()).Return(
					&compute.Command{
						Type: compute.CommandUSERS,
						Args: map[string]string{},
					}, nil).Once()
				us.On("ListUsernames", mock.Anything).Return([]string{"user1", "user2"}, nil).Once()
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
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				p.On("Parse", compute.CommandCREATEROLE.Make("role", "rwd", "namespace")).Return(
					&compute.Command{
						Type: compute.CommandCREATEROLE,
						Args: map[string]string{
							compute.RoleNameArg:    "role",
							compute.PermissionsArg: "rwd",
							compute.NamespaceArg:   "namespace",
						},
					}, nil).Once()
				ns.On("Exists", mock.Anything, "namespace").Return(true).Once()
				rs.On("Get", mock.Anything, "role").Return(nil, identity.ErrRoleNotFound).Once()
				rs.On("Save", mock.Anything, mock.MatchedBy(func(role *models.Role) bool {
					return role.Name == "role" && role.Namespace == "namespace" &&
						role.Get && role.Set && role.Del // Check parsed permissions
				})).Return(nil).Once()
				rs.On("Append", mock.Anything, "role").Return(nil, nil).Once()
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
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				p.On("Parse", compute.CommandDELETEROLE.Make("role")).Return(
					&compute.Command{
						Type: compute.CommandDELETEROLE,
						Args: map[string]string{
							compute.RoleNameArg: "role",
						},
					}, nil).Once()
				us.On("ListUsernames", mock.Anything).Return([]string{}, nil).Once()
				rs.On("Delete", mock.Anything, "role").Return(nil).Once()
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
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				p.On("Parse", compute.CommandROLES.String()).Return(
					&compute.Command{
						Type: compute.CommandROLES,
						Args: map[string]string{},
					}, nil).Once()
				rs.On("List", mock.Anything).Return([]string{"role1", "role2"}, nil).Once()
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
								Name:      "",
								Namespace: "default",
								Get:       true, Set: true, Del: true,
							},
						},
					}, nil).Once()
				p.On("Parse", compute.CommandME.String()).Return(
					&compute.Command{
						Type: compute.CommandME,
						Args: map[string]string{},
					}, nil).Once()
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
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				p.On("Parse", compute.CommandCREATENAMESPACE.Make("namespace")).Return(
					&compute.Command{
						Type: compute.CommandCREATENAMESPACE,
						Args: map[string]string{
							compute.NamespaceArg: "namespace",
						},
					}, nil).Once()
				ns.On("Save", mock.Anything, "namespace").Return(nil).Once()
				ns.On("Append", mock.Anything, "namespace").Return(nil, nil).Once()
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
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				rs.On("List", mock.Anything).Return(nil, nil)
				p.On("Parse", compute.CommandDELETENAMESPACE.Make("namespace")).Return(
					&compute.Command{
						Type: compute.CommandDELETENAMESPACE,
						Args: map[string]string{
							compute.NamespaceArg: "namespace",
						},
					}, nil).Once()
				ns.On("Delete", mock.Anything, "namespace").Return(nil).Once()
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
				userSession := &models.Session{
					User: &models.User{
						Username: "user",
						Roles:    []string{"role1"},
						ActiveRole: models.Role{
							Namespace: "default", Get: true,
						},
					},
				}
				ss.On("Get", sessionID).Return(userSession, nil).Once()
				p.On("Parse", compute.CommandSETNS.Make("namespace")).Return(
					&compute.Command{
						Type: compute.CommandSETNS,
						Args: map[string]string{
							compute.NamespaceArg: "namespace",
						},
					}, nil).Once()
				ns.On("Exists", mock.Anything, "namespace").Return(true).Once()
				rs.On("Get", mock.Anything, "role1").Return(&models.Role{
					Name: "role1", Namespace: "namespace", Get: true, Set: true,
				}, nil).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockParser := dbMock.NewParser(t)
			mockStorage := dbMock.NewStorage(t)
			mockUserStorage := dbMock.NewUsersStorage(t)
			mockRolesStorage := dbMock.NewRolesStorage(t)
			mockNamespaceStorage := dbMock.NewNamespacesStorage(t)
			mockSessionStorage := dbMock.NewSessionStorage(t)

			tt.prepareMocks(
				mockParser, mockStorage, mockUserStorage,
				mockNamespaceStorage, mockRolesStorage,
				mockSessionStorage,
			)

			db := New(
				mockParser,
				mockStorage,
				mockUserStorage,
				mockNamespaceStorage,
				mockRolesStorage,
				mockSessionStorage,
				&config.RootConfig{Username: "admin"},
			)

			result := db.HandleQuery(ctx, sessionID, tt.query)
			assert.Equal(t, tt.expected, result)

			mockParser.AssertExpectations(t)
			mockStorage.AssertExpectations(t)
			mockUserStorage.AssertExpectations(t)
			mockRolesStorage.AssertExpectations(t)
			mockNamespaceStorage.AssertExpectations(t)
			mockSessionStorage.AssertExpectations(t)
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
		expectedUser *models.User
		expectedErr  error
		prepareMocks func(p *dbMock.Parser, us *dbMock.UsersStorage, ss *dbMock.SessionStorage)
	}{
		{
			name:        "parse error",
			query:       "invalid query",
			expectedErr: fmt.Errorf("parse input failed: %w", errors.New("invalid command: unrecognized command")),
			prepareMocks: func(
				p *dbMock.Parser, us *dbMock.UsersStorage,
				ss *dbMock.SessionStorage,
			) {
				p.On("Parse", "invalid query").Return(nil, errors.New("invalid command: unrecognized command")).Once()
			},
		},
		{
			name:        "authentication failed (wrong password)",
			query:       compute.CommandAUTH.Make("username", "password"),
			expectedErr: errors.New("authentication failed"),
			prepareMocks: func(
				p *dbMock.Parser, us *dbMock.UsersStorage,
				ss *dbMock.SessionStorage,
			) {
				p.On("Parse", compute.CommandAUTH.Make("username", "password")).Return(
					&compute.Command{
						Type: compute.CommandAUTH,
						Args: map[string]string{
							compute.UsernameArg: "username",
							compute.PasswordArg: "password",
						},
					}, nil).Once()
				us.On("Authenticate", mock.Anything, "username", "password").Return(nil, identity.ErrAuthenticationFailed).Once()
			},
		},
		{
			name:        "authentication failed (user not found or internal error)",
			query:       compute.CommandAUTH.Make("username", "password"),
			expectedErr: ErrAuthenticationRequired,
			prepareMocks: func(
				p *dbMock.Parser, us *dbMock.UsersStorage,
				ss *dbMock.SessionStorage,
			) {
				p.On("Parse", compute.CommandAUTH.Make("username", "password")).Return(
					&compute.Command{
						Type: compute.CommandAUTH,
						Args: map[string]string{
							compute.UsernameArg: "username",
							compute.PasswordArg: "password",
						},
					}, nil).Once()
				us.On("Authenticate", mock.Anything, "username", "password").Return(nil, nil).Once()
			},
		},
		{
			name:         "successful login",
			query:        compute.CommandAUTH.Make("username", "password"),
			user:         &models.User{Username: "username"},
			expectedUser: &models.User{Username: "username"},
			expectedErr:  nil,
			prepareMocks: func(
				p *dbMock.Parser, us *dbMock.UsersStorage,
				ss *dbMock.SessionStorage,
			) {
				user := &models.User{Username: "username"}
				p.On("Parse", compute.CommandAUTH.Make("username", "password")).Return(
					&compute.Command{
						Type: compute.CommandAUTH,
						Args: map[string]string{
							compute.UsernameArg: "username",
							compute.PasswordArg: "password",
						},
					}, nil).Once()
				us.On("Authenticate", mock.Anything, "username", "password").Return(user, nil).Once()
				ss.On("Create", "1", user).Return(nil).Once()
			},
		},
		{
			name:        "login command missing password",
			query:       compute.CommandAUTH.String() + " username",
			expectedErr: fmt.Errorf("parse input failed: %w", fmt.Errorf("%w: missing required parameter '%s'", compute.ErrInvalidSyntax, compute.PasswordArg)),
			prepareMocks: func(
				p *dbMock.Parser, us *dbMock.UsersStorage,
				ss *dbMock.SessionStorage,
			) {
				parseError := fmt.Errorf("%w: missing required parameter '%s'", compute.ErrInvalidSyntax, compute.PasswordArg)
				p.On("Parse", compute.CommandAUTH.String()+" username").Return(nil, parseError).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			mockParser := dbMock.NewParser(t)
			mockUserStorage := dbMock.NewUsersStorage(t)
			mockSessionStorage := dbMock.NewSessionStorage(t)
			tt.prepareMocks(mockParser, mockUserStorage, mockSessionStorage)

			db := &Database{
				parser:      mockParser,
				userStorage: mockUserStorage,
				sessions:    mockSessionStorage,
			}

			user, err := db.Login(ctx, "1", tt.query)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.expectedErr.Error())
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, tt.expectedUser.Username, user.Username)
			}

			mockParser.AssertExpectations(t)
			mockUserStorage.AssertExpectations(t)
			mockSessionStorage.AssertExpectations(t)
		})
	}
}

func TestDatabase_Logout(t *testing.T) {
	t.Parallel()

	mockSessionStorage := dbMock.NewSessionStorage(t)
	db := &Database{sessions: mockSessionStorage}

	ctx := context.Background()
	sessionID := "1"

	mockSessionStorage.On("Delete", sessionID).Return(nil).Once()

	result := db.Logout(ctx, sessionID)
	assert.Equal(t, okPrefix, result)

	mockSessionStorage.AssertExpectations(t)
}
