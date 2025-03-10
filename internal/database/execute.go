package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

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
	session, err := db.sessions.Get(sessionID)
	if err != nil {
		logger.Debug("get current session failed", zap.Error(err), zap.String("session", sessionID))
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
		zap.Strings("args", cmd.Args),
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
		zap.Strings("args", cmd.Args),
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

// del - executes the del command to remove a key from the storage
func (db *Database) del(ctx context.Context, user *models.User, args []string) string {
	if !user.ActiveRole.Del {
		return WrapError(ErrPermissionDenied)
	}

	key := storage.MakeKey(user.ActiveRole.Namespace, args[0])
	if err := db.storage.Del(ctx, key); err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// get - executes the get command to retrieve the value of a key from the storage.
func (db *Database) get(ctx context.Context, user *models.User, args []string) string {
	if !user.ActiveRole.Get {
		return WrapError(ErrPermissionDenied)
	}

	key := storage.MakeKey(user.ActiveRole.Namespace, args[0])
	val, err := db.storage.Get(ctx, key)
	if err != nil {
		return WrapError(err)
	}

	return WrapOK(val)
}

// set - executes the SET command to store a key-value pair in the storage.
func (db *Database) set(ctx context.Context, user *models.User, args []string) string {
	if !user.ActiveRole.Set {
		return WrapError(ErrPermissionDenied)
	}

	key := storage.MakeKey(user.ActiveRole.Namespace, args[0])
	if err := db.storage.Set(ctx, key, args[1]); err != nil {
		return WrapError(err)
	}

	return okPrefix
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
		user, err = db.userStorage.Authenticate(ctx, cmd.Args[0], cmd.Args[1])
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
func (db *Database) help(ctx context.Context, usr *models.User, _ []string) string {
	if usr.IsAdmin(db.cfg) {
		return compute.AdminHelpText
	}

	return compute.UserHelpText
}

// help - executes the help command to print information about commands.
func (db *Database) listSessions(ctx context.Context, _ *models.User, _ []string) string {
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
func (db *Database) ns(ctx context.Context, _ *models.User, _ []string) string {
	namepaces, err := db.namespaceStorage.List(ctx)
	if err != nil {
		return WrapError(err)
	}

	// TODO: need to optimize
	res, err := json.Marshal(namepaces)
	if err != nil {
		return WrapError(err)
	}

	return WrapOK(string(res))
}

// createUser - executes the create user command to create a new user.
func (db *Database) createUser(ctx context.Context, _ *models.User, args []string) string {
	usr, err := db.userStorage.Create(ctx, args[0], args[1])
	if err != nil {
		return WrapError(err)
	}

	if _, err := db.userStorage.Append(ctx, usr.Username); err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// createUser - executes the create user command to create a new user.
func (db *Database) deleteUser(ctx context.Context, usr *models.User, args []string) string {
	if err := db.userStorage.Delete(ctx, args[0]); err != nil {
		return WrapError(err)
	}

	if _, err := db.userStorage.Remove(ctx, usr.Username); err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// assignRole - executes the assign role command to assign a role to a user.
func (db *Database) assignRole(ctx context.Context, _ *models.User, args []string) string {
	if err := db.userStorage.AssignRole(ctx, args[0], args[1]); err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// assignRole - executes the assign role command to assign a role to a user.
func (db *Database) divestRole(ctx context.Context, _ *models.User, args []string) string {
	if err := db.userStorage.AssignRole(ctx, args[0], args[1]); err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// users - executes the users command to list all usernames.
func (db *Database) users(ctx context.Context, _ *models.User, args []string) string {
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

func (db *Database) getUser(ctx context.Context, _ *models.User, args []string) string {
	user, err := db.userStorage.Get(ctx, args[0])
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

func (db *Database) getRole(ctx context.Context, _ *models.User, args []string) string {
	role, err := db.rolesStorage.Get(ctx, args[0])
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
func (db *Database) createRole(ctx context.Context, _ *models.User, args []string) string {
	if !db.namespaceStorage.Exists(ctx, args[2]) {
		return WrapError(identity.ErrNamespaceNotFound)
	}

	_, err := db.rolesStorage.Get(ctx, args[0])
	if err != nil && !errors.Is(err, identity.ErrRoleNotFound) {
		return WrapError(err)
	}

	role, err := models.NewRole(args[0], args[1], args[2])
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
func (db *Database) delRole(ctx context.Context, _ *models.User, args []string) string {
	users, err := db.userStorage.ListUsernames(ctx)
	if err != nil && !errors.Is(err, identity.ErrEmptyUsers) {
		return WrapError(err)
	}

	for _, username := range users {
		user, err := db.userStorage.Get(ctx, username)
		if err != nil {
			return WrapError(err)
		}

		if slices.Contains(user.Roles, args[0]) {
			return WrapError(fmt.Errorf("cannot delete assigned to user '%s' role", username))
		}
	}

	if err := db.rolesStorage.Delete(ctx, args[0]); err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// listRoles - executes the listRoles command to list all listRoles.
func (db *Database) listRoles(ctx context.Context, _ *models.User, args []string) string {
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
func (db *Database) me(ctx context.Context, user *models.User, _ []string) string {
	return WrapOK(fmt.Sprintf(
		"user: '%s', roles: '%s', ns: '%s', perms: '%s'",
		user.Username, strings.Join(user.Roles, " "),
		user.ActiveRole.Namespace, user.ActiveRole.Perms(),
	))
}

// createNS - executes the create ns command to create a new namespace.
func (db *Database) createNS(ctx context.Context, _ *models.User, args []string) string {
	err := db.namespaceStorage.Save(ctx, args[0])
	if err != nil {
		return WrapError(err)
	}

	if _, err = db.namespaceStorage.Append(ctx, args[0]); err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// deleteNS - executes the delete ns command to delete a namespace
func (db *Database) deleteNS(ctx context.Context, _ *models.User, args []string) string {
	roles, err := db.rolesStorage.List(ctx)
	if err != nil {
		return WrapError(err)
	}

	for _, name := range roles {
		role, err := db.rolesStorage.Get(ctx, name)
		if err != nil {
			continue
		}

		if role.Namespace == args[0] {
			return WrapError(
				fmt.Errorf(
					"this namespace is still used by the role %s",
					role.Name),
			)
		}
	}

	err = db.namespaceStorage.Delete(ctx, args[0])
	if err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// setNamespace - executes the set ns command to change the current namespace for the user.
func (db *Database) setNamespace(ctx context.Context, user *models.User, args []string) string {
	namespace := args[0]

	if !db.namespaceStorage.Exists(ctx, namespace) {
		return WrapError(identity.ErrNamespaceNotFound)
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
			return WrapError(ErrPermissionDenied)
		}
	}

	if user.IsAdmin(db.cfg) {
		role = &models.Role{
			Get: true, Set: true, Del: true,
			Namespace: namespace,
		}
	}

	user.ActiveRole = *role
	return okPrefix
}

// watch - watches the key and returns the value if it has changed.
func (db *Database) watch(ctx context.Context, user *models.User, args []string) string {
	key := storage.MakeKey(user.ActiveRole.Namespace, args[0])
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
