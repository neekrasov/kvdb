package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/neekrasov/kvdb/internal/database/compute"
	"github.com/neekrasov/kvdb/internal/database/identity"
	"github.com/neekrasov/kvdb/internal/database/identity/models"
	"github.com/neekrasov/kvdb/internal/database/storage"
	"github.com/neekrasov/kvdb/pkg/ctxutil"
	"github.com/neekrasov/kvdb/pkg/logger"
	"go.uber.org/zap"
)

var (
	ErrInvalidOperation       = errors.New("invalid operation")
	ErrAuthenticationRequired = errors.New("authentication required")
	ErrPermissionDenied       = errors.New("permission denied")
	ErrEmptyResult            = errors.New("empty result")
)

// HandleQuery -  processes a user query by parsing and executing the corresponding command.
func (db *Database) HandleQuery(ctx context.Context, sessionID string, query string) string {
	if ctx.Err() != nil {
		return WrapError(ctx.Err())
	}

	session, err := db.sessions.Get(sessionID)
	if err != nil {
		logger.Debug("get current session failed", zap.Error(err),
			zap.String("session", sessionID))
		return WrapError(fmt.Errorf("get current session failed: %w", err))
	}

	cmd, err := db.parser.Parse(query)
	if err != nil {
		logger.Debug(
			"parse query failed", zap.Error(err),
			zap.String("session", sessionID),
		)
		return WrapError(fmt.Errorf("parse input failed: %w", err))
	}

	logger.Info("parsed command",
		zap.Stringer("cmd_type", cmd.Type),
		zap.Any("args", cmd.Args),
		zap.String("session", sessionID),
	)

	handler, ok := db.registry[cmd.Type]
	if !ok {
		return WrapError(ErrInvalidOperation)
	}

	if handler.AdminOnly && session.User.Username != db.cfg.Username {
		return WrapError(ErrPermissionDenied)
	}

	oldUsr := session.User
	ctx = ctxutil.InjectSessionID(ctx, sessionID)
	result := handler.Func(ctx, session.User, cmd.Args)
	logger.Info("operation executed",
		zap.Stringer("cmd_type", cmd.Type),
		zap.Any("args", cmd.Args),
		zap.String("result", result),
		zap.Bool("error", IsError(result)),
		zap.String("session", sessionID))

	if oldUsr != session.User {
		err := db.userStorage.SaveRaw(ctx, session.User)
		logger.Debug(
			"save changed user failed",
			zap.Any("old_user", oldUsr),
			zap.Any("new_user", session.User),
			zap.String("session", sessionID),
			zap.Error(err),
		)
	}

	return result

}

// Login - authenticates a user based on the provided query.
func (db *Database) Login(ctx context.Context, sessionID string, query string) (*models.User, error) {
	cmd, err := db.parser.Parse(query)
	if err != nil {
		logger.Debug("parse query failed", zap.Error(err))
		return nil, fmt.Errorf("parse input failed: %w", err)
	}

	var user *models.User
	if cmd.Type == compute.CommandAUTH {
		username := cmd.Args[compute.UsernameArg]
		password := cmd.Args[compute.PasswordArg]

		user, err = db.userStorage.Authenticate(ctx, username, password)
		if err != nil {
			return nil, err
		}
	}

	if user == nil {
		return nil, ErrAuthenticationRequired
	}

	if err = db.sessions.Create(sessionID, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Logout - logs out the user by deleting their session token.
func (db *Database) Logout(ctx context.Context, sessionID string) string {
	db.sessions.Delete(sessionID)

	return okPrefix
}

// help - executes the help command to print information about commands.
func (db *Database) help(ctx context.Context, usr *models.User, _ Args) string {
	if usr.IsAdmin(db.cfg) {
		return WrapOK(compute.AdminHelpText)
	}

	return WrapOK(compute.UserHelpText)
}

// del - executes the del command to remove a key from the storage
func (db *Database) del(ctx context.Context, user *models.User, args Args) string {
	namespace, err := db.parseNS(ctx, user, args)
	if err != nil {
		return WrapError(err)
	}

	role := db.checkPermissions(ctx, user, namespace)
	if role == nil || !role.Del {
		return WrapError(ErrPermissionDenied)
	}

	key := storage.MakeKey(namespace, args["key"])
	if err := db.storage.Del(ctx, key); err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// get - executes the get command to retrieve the value of a key from the storage.
func (db *Database) get(ctx context.Context, user *models.User, args Args) string {
	namespace, err := db.parseNS(ctx, user, args)
	if err != nil {
		return WrapError(err)
	}

	role := db.checkPermissions(ctx, user, namespace)
	if role == nil || !role.Get {
		return WrapError(ErrPermissionDenied)
	}

	key := storage.MakeKey(namespace, args["key"])
	val, err := db.storage.Get(ctx, key)
	if err != nil {
		return WrapError(err)
	}

	return WrapOK(val)
}

// set - executes the SET command to store a key-value pair in the storage.
func (db *Database) set(ctx context.Context, user *models.User, args Args) string {
	namespace, err := db.parseNS(ctx, user, args)
	if err != nil {
		return WrapError(err)
	}

	role := db.checkPermissions(ctx, user, namespace)
	if role == nil || !role.Set {
		return WrapError(ErrPermissionDenied)
	}

	if val, ok := args[compute.TTLArg]; ok {
		ctx = ctxutil.InjectTTL(ctx, val)
	}

	key := storage.MakeKey(namespace, args[compute.KeyArg])
	if err := db.storage.Set(ctx, key, args[compute.ValueArg]); err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// help - executes the help command to print information about commands.
func (db *Database) listSessions(ctx context.Context, _ *models.User, _ Args) string {
	sessions := db.sessions.List()
	if len(sessions) == 0 {
		return WrapError(ErrEmptyResult)
	}

	// TODO: need to optimize
	res, err := json.Marshal(sessions)
	if err != nil {
		return WrapError(err)
	}

	return WrapOK(string(res))
}

// ns - executes the ns command to list namespaces.
func (db *Database) ns(ctx context.Context, user *models.User, _ Args) string {
	namepaces, err := db.namespaceStorage.List(ctx)
	if err != nil {
		return WrapError(err)
	}

	var res []byte
	if user.IsAdmin(db.cfg) {
		// TODO: need to optimize
		res, err = json.Marshal(namepaces)
		if err != nil {
			return WrapError(err)
		}
	} else {
		userNamespaces := make([]string, 0, len(user.Roles))
		for _, namespace := range namepaces {
			for _, roleName := range user.Roles {
				role, err := db.rolesStorage.Get(ctx, roleName)
				if err != nil {
					continue
				}
				if role.Namespace == namespace {
					userNamespaces = append(userNamespaces, role.Namespace)
				}
			}
		}

		// TODO: need to optimize
		res, err = json.Marshal(userNamespaces)
		if err != nil {
			return WrapError(err)
		}
	}

	return WrapOK(string(res))
}

// createUser - executes the create user command to create a new user.
func (db *Database) createUser(ctx context.Context, _ *models.User, args Args) string {
	username := args[compute.UsernameArg]
	password := args[compute.PasswordArg]

	usr, err := db.userStorage.Create(ctx, username, password)
	if err != nil {
		return WrapError(err)
	}

	if _, err := db.userStorage.Append(ctx, usr.Username); err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// createUser - executes the create user command to create a new user.
func (db *Database) deleteUser(ctx context.Context, usr *models.User, args Args) string {
	username := args[compute.UsernameArg]

	if err := db.userStorage.Delete(ctx, username); err != nil {
		return WrapError(err)
	}

	if _, err := db.userStorage.Remove(ctx, usr.Username); err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// assignRole - executes the assign role command to assign a role to a user.
func (db *Database) assignRole(ctx context.Context, _ *models.User, args Args) string {
	username := args[compute.UsernameArg]
	role := args[compute.RoleArg]

	if err := db.userStorage.AssignRole(ctx, username, role); err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// assignRole - executes the assign role command to assign a role to a user.
func (db *Database) divestRole(ctx context.Context, _ *models.User, args Args) string {
	username := args[compute.UsernameArg]
	role := args[compute.RoleArg]

	if err := db.userStorage.AssignRole(ctx, username, role); err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// users - executes the users command to list all usernames.
func (db *Database) users(ctx context.Context, _ *models.User, _ Args) string {
	users, err := db.userStorage.ListUsernames(ctx)
	if err != nil {
		return WrapError(err)
	}

	if len(users) == 0 {
		return WrapError(ErrEmptyResult)
	}

	// TODO: need to optimize
	res, err := json.Marshal(users)
	if err != nil {
		return WrapError(err)
	}

	return WrapOK(string(res))
}

func (db *Database) getUser(ctx context.Context, _ *models.User, args Args) string {
	username := args[compute.UsernameArg]

	user, err := db.userStorage.Get(ctx, username)
	if err != nil {
		return WrapError(err)
	}

	// TODO: need to optimize
	res, err := json.Marshal(user)
	if err != nil {
		return WrapError(err)
	}

	return WrapOK(string(res))
}

func (db *Database) getRole(ctx context.Context, _ *models.User, args Args) string {
	roleName := args[compute.RoleNameArg]

	role, err := db.rolesStorage.Get(ctx, roleName)
	if err != nil {
		return WrapError(err)
	}

	// TODO: need to optimize
	res, err := json.Marshal(role)
	if err != nil {
		return WrapError(err)
	}

	return WrapOK(string(res))
}

// createRole - executes the create role command to create a new role.
func (db *Database) createRole(ctx context.Context, _ *models.User, args Args) string {
	namespace := args[compute.NamespaceArg]
	roleName := args[compute.RoleNameArg]
	permissions := args[compute.PermissionsArg]

	if !db.namespaceStorage.Exists(ctx, namespace) {
		return WrapError(identity.ErrNamespaceNotFound)
	}

	_, err := db.rolesStorage.Get(ctx, roleName)
	if err != nil && !errors.Is(err, identity.ErrRoleNotFound) {
		return WrapError(err)
	}

	role, err := models.NewRole(roleName, permissions, namespace)
	if err != nil {
		return WrapError(err)
	}

	if err := db.rolesStorage.Save(ctx, &role); err != nil {
		return WrapError(err)
	}

	if _, err := db.rolesStorage.Append(ctx, role.Name); err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// delRole - executes command to delete a role.
func (db *Database) delRole(ctx context.Context, _ *models.User, args Args) string {
	roleName := args[compute.RoleNameArg]

	users, err := db.userStorage.ListUsernames(ctx)
	if err != nil && !errors.Is(err, identity.ErrEmptyUsers) {
		return WrapError(err)
	}

	for _, username := range users {
		user, err := db.userStorage.Get(ctx, username)
		if err != nil {
			return WrapError(err)
		}

		if slices.Contains(user.Roles, roleName) {
			return WrapError(fmt.Errorf("cannot delete assigned to user '%s' role", username))
		}
	}

	if err := db.rolesStorage.Delete(ctx, roleName); err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// listRoles - executes the listRoles command to list all listRoles.
func (db *Database) listRoles(ctx context.Context, _ *models.User, _ Args) string {
	roles, err := db.rolesStorage.List(ctx)
	if err != nil {
		return WrapError(err)
	}

	if len(roles) == 0 {
		return WrapError(ErrEmptyResult)
	}

	// TODO: need to optimize
	res, err := json.Marshal(roles)
	if err != nil {
		return WrapError(err)
	}

	return WrapOK(string(res))
}

// me - executes the me command to display information about the current user,
// including their username, roles, namespace, and permissions.
func (db *Database) me(ctx context.Context, user *models.User, _ Args) string {
	return WrapOK(fmt.Sprintf(
		"user: '%s', roles: '%s', ns: '%s', perms: '%s'",
		user.Username, strings.Join(user.Roles, " "),
		user.ActiveRole.Namespace, user.ActiveRole.Perms(),
	))
}

// createNS - executes the create ns command to create a new namespace.
func (db *Database) createNS(ctx context.Context, _ *models.User, args Args) string {
	namespace := args[compute.NamespaceArg]

	err := db.namespaceStorage.Save(ctx, namespace)
	if err != nil {
		return WrapError(err)
	}

	if _, err = db.namespaceStorage.Append(ctx, namespace); err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// deleteNS - executes the delete ns command to delete a namespace
func (db *Database) deleteNS(ctx context.Context, _ *models.User, args Args) string {
	roles, err := db.rolesStorage.List(ctx)
	if err != nil {
		return WrapError(err)
	}

	namespace := args[compute.NamespaceArg]

	for _, name := range roles {
		role, err := db.rolesStorage.Get(ctx, name)
		if err != nil {
			continue
		}

		if role.Namespace == namespace {
			return WrapError(
				fmt.Errorf(
					"this namespace is still used by the role %s",
					role.Name),
			)
		}
	}

	err = db.namespaceStorage.Delete(ctx, namespace)
	if err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// setNamespace - executes the set ns command to change the current namespace for the user.
func (db *Database) setNamespace(ctx context.Context, user *models.User, args Args) string {
	namespace := args[compute.NamespaceArg]

	if !db.namespaceStorage.Exists(ctx, namespace) {
		return WrapError(identity.ErrNamespaceNotFound)
	}

	role := db.checkPermissions(ctx, user, namespace)
	if role == nil {
		return WrapError(ErrPermissionDenied)
	}

	user.ActiveRole = *role
	return okPrefix
}

// watch - watches the key and returns the value if it has changed.
func (db *Database) watch(ctx context.Context, user *models.User, args Args) string {
	namespace, err := db.parseNS(ctx, user, args)
	if err != nil {
		return WrapError(err)
	}

	role := db.checkPermissions(ctx, user, namespace)
	if role == nil || !role.Get {
		return WrapError(ErrPermissionDenied)
	}

	key := storage.MakeKey(namespace, args[compute.KeyArg])
	future := db.storage.Watch(ctx, key)

	ch := make(chan string)
	go func() {
		ch <- future.Get()
	}()

	for {
		select {
		case <-ctx.Done():
			return okPrefix
		case val := <-ch:
			return WrapOK(val)
		}
	}
}

// stat - displays database statistics.
func (db *Database) stat(ctx context.Context, _ *models.User, _ Args) string {
	storageStats, err := db.storage.Stats()
	if err != nil {
		return WrapError(err)
	}

	namespaces, err := db.namespaceStorage.List(ctx)
	if err != nil {
		return WrapError(err)
	}

	roles, err := db.rolesStorage.List(ctx)
	if err != nil {
		return WrapError(err)
	}

	users, err := db.userStorage.ListUsernames(ctx)
	if err != nil {
		return WrapError(err)
	}

	stats := Stats{
		ActiveSessions:  int64(len(db.sessions.List())),
		TotalNamespaces: int64(len(namespaces)),
		TotalRoles:      int64(len(roles)),
		TotalUsers:      int64(len(users)),
		Uptime:          time.Since(storageStats.StartTime).Seconds(),
		TotalCommands:   storageStats.TotalCommands.Load(),
		GetCommands:     storageStats.GetCommands.Load(),
		SetCommands:     storageStats.SetCommands.Load(),
		DelCommands:     storageStats.DelCommands.Load(),
		TotalKeys:       storageStats.TotalKeys.Load(),
		ExpiredKeys:     storageStats.ExpiredKeys.Load(),
	}

	res, err := json.Marshal(stats)
	if err != nil {
		return WrapError(err)
	}

	return WrapOK(string(res))
}

func (db *Database) parseNS(ctx context.Context, user *models.User, args Args) (string, error) {
	var namespace string
	if val, ok := args[compute.NSArg]; ok {
		if !db.namespaceStorage.Exists(ctx, val) {
			return "", identity.ErrNamespaceNotFound
		}

		namespace = val
	} else {
		namespace = user.ActiveRole.Namespace
	}

	return namespace, nil
}

func (db *Database) checkPermissions(
	ctx context.Context, user *models.User, namespace string,
) *models.Role {
	if user.ActiveRole.Namespace == namespace {
		return &user.ActiveRole
	}

	var (
		role      *models.Role
		hasAccess bool
	)
	if !user.IsAdmin(db.cfg) {
		for _, roleName := range user.Roles {
			var err error
			role, err = db.rolesStorage.Get(ctx, roleName)
			if err != nil {
				continue
			}
			if role.Namespace == namespace {
				hasAccess = true
				break
			}
		}

		if !hasAccess {
			return nil
		}
	} else {
		role = &models.Role{
			Get: true, Set: true, Del: true,
			Namespace: namespace,
		}
	}

	return role
}
