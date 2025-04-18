package database

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/neekrasov/kvdb/internal/config"
	"github.com/neekrasov/kvdb/internal/database/compute"
	"github.com/neekrasov/kvdb/internal/database/identity"
	"github.com/neekrasov/kvdb/internal/database/identity/models"
	"github.com/neekrasov/kvdb/internal/database/storage"
	dbMock "github.com/neekrasov/kvdb/internal/mocks/database"
	"github.com/neekrasov/kvdb/pkg/logger"
	pkgsync "github.com/neekrasov/kvdb/pkg/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestDatabase_HandleQuery(t *testing.T) {
	t.Parallel()
	logger.MockLogger()

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte("password"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	sessionID := "test"
	adminSession := &models.Session{
		User: &models.User{
			Username: "admin",
			Password: string(hashedPassword),
			ActiveRole: models.Role{
				Get:       true,
				Set:       true,
				Del:       true,
				Namespace: models.DefaultNameSpace,
			},
		},
	}

	userActiveRole := models.Role{
		Name:      "role1",
		Namespace: "default", Get: true,
	}

	userSession := &models.Session{
		User: &models.User{
			Username:   "user",
			Roles:      []string{"role1"},
			ActiveRole: userActiveRole,
		},
	}

	tests := []struct {
		name         string
		query        string
		parseErr     error
		expected     string
		contains     string
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
						role.Get && role.Set && role.Del
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
				us.On("ListUsernames", mock.Anything).Return([]string{"test"}, nil).Once()
				us.On("Get", mock.Anything, "test").Return(&models.User{Username: "test", Roles: []string{}}, nil).Once()
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
		{
			name:     "delete user command",
			query:    compute.CommandDELETEUSER.Make("username"),
			expected: okPrefix,
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				p.On("Parse", compute.CommandDELETEUSER.Make("username")).Return(
					&compute.Command{
						Type: compute.CommandDELETEUSER,
						Args: map[string]string{
							compute.UsernameArg: "username",
						},
					}, nil).Once()
				us.On("Delete", mock.Anything, "username").Return(nil).Once()
				us.On("Remove", mock.Anything, "admin").Return(nil, nil).Once()
			},
		},
		{
			name:     "divest role command",
			query:    compute.CommandDIVESTROLE.Make("username", "role"),
			expected: okPrefix,
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				p.On("Parse", compute.CommandDIVESTROLE.Make("username", "role")).Return(
					&compute.Command{
						Type: compute.CommandDIVESTROLE,
						Args: map[string]string{
							compute.UsernameArg: "username",
							compute.RoleArg:     "role",
						},
					}, nil).Once()
				us.On("AssignRole", mock.Anything, "username", "role").Return(nil).Once()
			},
		},
		{
			name:     "get user command",
			query:    compute.CommandGETUSER.Make("username"),
			expected: okPrefix + ` {"username":"username","roles":["role1"],"role":{"name":"","get":false,"set":false,"del":false,"namespace":""}}`,
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				p.On("Parse", compute.CommandGETUSER.Make("username")).Return(
					&compute.Command{
						Type: compute.CommandGETUSER,
						Args: map[string]string{
							compute.UsernameArg: "username",
						},
					}, nil).Once()
				us.On("Get", mock.Anything, "username").Return(&models.User{
					Username: "username",
					Roles:    []string{"role1"},
				}, nil).Once()
			},
		},
		{
			name:     "get role command",
			query:    compute.CommandGETROLE.Make("role"),
			expected: okPrefix + ` {"name":"role","get":true,"set":true,"del":true,"namespace":"default"}`,
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				p.On("Parse", compute.CommandGETROLE.Make("role")).Return(
					&compute.Command{
						Type: compute.CommandGETROLE,
						Args: map[string]string{
							compute.RoleNameArg: "role",
						},
					}, nil).Once()
				rs.On("Get", mock.Anything, "role").Return(&models.Role{
					Name:      "role",
					Namespace: "default",
					Get:       true,
					Set:       true,
					Del:       true,
				}, nil).Once()
			},
		},
		{
			name:     "watch command success",
			query:    compute.CommandWATCH.Make("key"),
			expected: okPrefix + " new_value",
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				p.On("Parse", compute.CommandWATCH.Make("key")).Return(
					&compute.Command{
						Type: compute.CommandWATCH,
						Args: map[string]string{
							compute.KeyArg: "key",
						},
					}, nil).Once()
				future := pkgsync.NewFuture[string]()
				go future.Set("new_value")

				s.On("Watch", mock.Anything, "default:key").Return(future).Once()
			},
		},
		{
			name:     "stat command success",
			query:    compute.CommandSTAT.String(),
			contains: `total_commands":100,"get_commands":50,"set_commands":30,"del_commands":20,"total_keys":1000,"expired_keys":50,"active_sessions":1,"total_namespaces":2,"total_roles":3,"total_users":4}`,
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				p.On("Parse", compute.CommandSTAT.String()).Return(
					&compute.Command{
						Type: compute.CommandSTAT,
						Args: map[string]string{},
					}, nil).Once()
				stats := &storage.Stats{}
				stats.TotalCommands.Store(100)
				stats.GetCommands.Store(50)
				stats.SetCommands.Store(30)
				stats.DelCommands.Store(20)
				stats.TotalKeys.Store(1000)
				stats.ExpiredKeys.Store(50)
				stats.StartTime, _ = time.Parse(time.RFC3339Nano, "2025-04-14T00:23:29.042785+03:00")
				s.On("Stats").Return(stats, nil).Once()
				ns.On("List", mock.Anything).Return([]string{"ns1", "ns2"}, nil).Once()
				rs.On("List", mock.Anything).Return([]string{"r1", "r2", "r3"}, nil).Once()
				us.On("ListUsernames", mock.Anything).Return([]string{"u1", "u2", "u3", "u4"}, nil).Once()
				ss.On("List").Return([]models.Session{
					{User: nil, ExpiresAt: time.Now(), CreatedAt: time.Now()},
				}).Once()
			},
		},
		{
			name:     "list sessions command",
			query:    compute.CommandSESSIONS.String(),
			expected: okPrefix + ` [{"user":null,"expires_at":"2025-04-14T00:23:29.042785+03:00","created_at":"2025-04-14T00:23:29.042785+03:00"}]`,
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				p.On("Parse", compute.CommandSESSIONS.String()).Return(
					&compute.Command{
						Type: compute.CommandSESSIONS,
						Args: map[string]string{},
					}, nil).Once()

				expiresAt, _ := time.Parse(time.RFC3339Nano, "2025-04-14T00:23:29.042785+03:00")
				createdAt, _ := time.Parse(time.RFC3339Nano, "2025-04-14T00:23:29.042785+03:00")

				ss.On("List").Return([]models.Session{
					{User: nil, ExpiresAt: expiresAt, CreatedAt: createdAt},
				}).Once()
			},
		},
		{
			name:     "namespace not found error",
			query:    compute.CommandSET.Make("key", "value") + " NS notfound",
			expected: fmt.Sprintf("%s namespace not found", errPrefix),
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				p.On("Parse", compute.CommandSET.Make("key", "value")+" NS notfound").Return(
					&compute.Command{
						Type: compute.CommandSET,
						Args: map[string]string{
							compute.KeyArg:   "key",
							compute.ValueArg: "value",
							compute.NSArg:    "notfound",
						},
					}, nil).Once()
				ns.On("Exists", mock.Anything, "notfound").Return(false).Once()
			},
		},
		{
			name:     "role not found error when setting namespace",
			query:    compute.CommandSETNS.Make("namespace"),
			expected: fmt.Sprintf("%s permission denied", errPrefix),
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
				rs.On("Get", mock.Anything, "role1").Return(nil, identity.ErrRoleNotFound).Once()
			},
		},
		{
			name:     "empty result for users command",
			query:    compute.CommandUSERS.String(),
			expected: fmt.Sprintf("%s empty result", errPrefix),
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
				us.On("ListUsernames", mock.Anything).Return([]string{}, nil).Once()
			},
		},
		{
			name:     "error listing users",
			query:    compute.CommandUSERS.String(),
			expected: fmt.Sprintf("%s internal error", errPrefix),
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
				us.On("ListUsernames", mock.Anything).Return(nil, errors.New("internal error")).Once()
			},
		},
		{
			name:     "error creating namespace",
			query:    compute.CommandCREATENAMESPACE.Make("namespace"),
			expected: fmt.Sprintf("%s internal error", errPrefix),
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
				ns.On("Save", mock.Anything, "namespace").Return(errors.New("internal error")).Once()
			},
		},
		{
			name:     "error deleting namespace",
			query:    compute.CommandDELETENAMESPACE.Make("namespace"),
			expected: fmt.Sprintf("%s internal error", errPrefix),
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
				ns.On("Delete", mock.Anything, "namespace").Return(errors.New("internal error")).Once()
			},
		},
		{
			name:     "namespace in use when deleting",
			query:    compute.CommandDELETENAMESPACE.Make("namespace"),
			expected: fmt.Sprintf("%s this namespace is still used by the role role1", errPrefix),
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				rs.On("List", mock.Anything).Return([]string{"role1"}, nil)
				p.On("Parse", compute.CommandDELETENAMESPACE.Make("namespace")).Return(
					&compute.Command{
						Type: compute.CommandDELETENAMESPACE,
						Args: map[string]string{
							compute.NamespaceArg: "namespace",
						},
					}, nil).Once()
				rs.On("Get", mock.Anything, "role1").Return(&models.Role{
					Name: "role1", Namespace: "namespace",
				}, nil).Once()
			},
		},
		{
			name:     "successful listing namespaces",
			query:    compute.CommandNAMESPACES.String(),
			expected: okPrefix + ` ["ns1","ns2","ns3"]`,
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				p.On("Parse", compute.CommandNAMESPACES.String()).Return(
					&compute.Command{
						Type: compute.CommandNAMESPACES,
						Args: map[string]string{},
					}, nil).Once()
				ns.On("List", mock.Anything).Return([]string{"ns1", "ns2", "ns3"}, nil).Once()
			},
		},
		{
			name:     "successful listing namespaces",
			query:    compute.CommandNAMESPACES.String(),
			expected: okPrefix + ` ["default"]`,
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(userSession, nil).Once()
				p.On("Parse", compute.CommandNAMESPACES.String()).Return(
					&compute.Command{
						Type: compute.CommandNAMESPACES,
						Args: map[string]string{},
					}, nil).Once()
				rs.On("Get", mock.Anything, "role1").Return(&userActiveRole, nil).Once()
			},
		},
		{
			name:     "error listing namespaces",
			query:    compute.CommandNAMESPACES.String(),
			expected: fmt.Sprintf("%s internal error", errPrefix),
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				p.On("Parse", compute.CommandNAMESPACES.String()).Return(
					&compute.Command{
						Type: compute.CommandNAMESPACES,
						Args: map[string]string{},
					}, nil).Once()
				ns.On("List", mock.Anything).Return(nil, errors.New("internal error")).Once()
			},
		},
		{
			name:     "error creating user",
			query:    compute.CommandCREATEUSER.Make("username", "password"),
			expected: fmt.Sprintf("%s internal error", errPrefix),
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
				us.On("Create", mock.Anything, "username", "password").Return(nil, errors.New("internal error")).Once()
			},
		},
		{
			name:     "error assigning role",
			query:    compute.CommandASSIGNROLE.Make("username", "role"),
			expected: fmt.Sprintf("%s internal error", errPrefix),
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
				us.On("AssignRole", mock.Anything, "username", "role").Return(errors.New("internal error")).Once()
			},
		},
		{
			name:     "error getting user",
			query:    compute.CommandGETUSER.Make("username"),
			expected: fmt.Sprintf("%s internal error", errPrefix),
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				p.On("Parse", compute.CommandGETUSER.Make("username")).Return(
					&compute.Command{
						Type: compute.CommandGETUSER,
						Args: map[string]string{
							compute.UsernameArg: "username",
						},
					}, nil).Once()
				us.On("Get", mock.Anything, "username").Return(nil, errors.New("internal error")).Once()
			},
		},
		{
			name:     "getting admin help",
			query:    compute.CommandHELP.Make(),
			expected: fmt.Sprintf("%s %s", okPrefix, compute.AdminHelpText),
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(adminSession, nil).Once()
				p.On("Parse", compute.CommandHELP.Make()).Return(
					&compute.Command{
						Type: compute.CommandHELP,
						Args: map[string]string{},
					}, nil).Once()
			},
		},
		{
			name:     "getting user help",
			query:    compute.CommandHELP.Make(),
			expected: fmt.Sprintf("%s %s", okPrefix, compute.UserHelpText),
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(userSession, nil).Once()
				p.On("Parse", compute.CommandHELP.Make()).Return(
					&compute.Command{
						Type: compute.CommandHELP,
						Args: map[string]string{},
					}, nil).Once()
			},
		},
		{
			name:     "error getting current session",
			query:    compute.CommandHELP.Make(),
			expected: fmt.Sprintf("%s get current session failed: test", errPrefix),
			prepareMocks: func(
				p *dbMock.Parser, s *dbMock.Storage,
				us *dbMock.UsersStorage, ns *dbMock.NamespacesStorage,
				rs *dbMock.RolesStorage, ss *dbMock.SessionStorage,
			) {
				ss.On("Get", sessionID).Return(nil, errors.New("test")).Once()
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
				&config.RootConfig{Username: "admin", Password: "password"},
			)

			result := db.HandleQuery(ctx, sessionID, tt.query)
			if tt.expected == "" {
				assert.Contains(t, result, tt.contains)
			} else {
				assert.Equal(t, tt.expected, result)
			}

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
