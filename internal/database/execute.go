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
func (c *Database) HandleQuery(ctx context.Context, sessionID string, query string) string {
	session, err := c.sessions.Get(sessionID)
	if err != nil {
		logger.Debug("get current session failed", zap.Error(err), zap.String("session_id", sessionID))
		return WrapError(fmt.Errorf("get current session failed: %w", err))
	}

	cmd, err := c.parser.Parse(query)
	if err != nil {
		logger.Debug(
			"parse query failed", zap.Error(err),
			zap.String("session_id", sessionID),
		)
		return WrapError(fmt.Errorf("parse input failed: %w", err))
	}

	logger.Info("parsed command",
		zap.Stringer("cmd_type", cmd.Type),
		zap.Strings("args", cmd.Args),
		zap.String("session_id", sessionID),
	)

	handler, ok := map[compute.CommandType]CommandHandler{
		compute.CommandCREATEUSER:      {Func: c.createUser, AdminOnly: true},
		compute.CommandASSIGNROLE:      {Func: c.assignRole, AdminOnly: true},
		compute.CommandCREATEROLE:      {Func: c.createRole, AdminOnly: true},
		compute.CommandDELETEROLE:      {Func: c.delRole, AdminOnly: true},
		compute.CommandROLES:           {Func: c.listRoles, AdminOnly: true},
		compute.CommandGETROLE:         {Func: c.getRole, AdminOnly: true},
		compute.CommandUSERS:           {Func: c.users, AdminOnly: true},
		compute.CommandGETUSER:         {Func: c.getUser, AdminOnly: true},
		compute.CommandCREATENAMESPACE: {Func: c.createNS, AdminOnly: true},
		compute.CommandDELETENAMESPACE: {Func: c.deleteNS, AdminOnly: true},
		compute.CommandNAMESPACES:      {Func: c.ns, AdminOnly: true},
		compute.CommandSESSIONS:        {Func: c.listSessions, AdminOnly: true},
		compute.CommandDELETEUSER:      {Func: c.deleteUser, AdminOnly: true},
		compute.CommandDIVESTROLE:      {Func: c.divestRole, AdminOnly: true},
		compute.CommandHELP:            {Func: c.help},
		compute.CommandSETNS:           {Func: c.setNamespace},
		compute.CommandME:              {Func: c.me},
		compute.CommandGET:             {Func: c.get},
		compute.CommandSET:             {Func: c.set},
		compute.CommandDEL:             {Func: c.del},
	}[cmd.Type]

	if !ok {
		return WrapError(ErrInvalidOperation)
	}

	if handler.AdminOnly && session.User.Username != c.cfg.Username {
		return WrapError(ErrPermissionDenied)
	}

	oldUsr := session.User

	// TODO: inject session_id in context.
	result := handler.Func(ctx, session.User, cmd.Args)
	logger.Info("operation executed",
		zap.Stringer("cmd_type", cmd.Type),
		zap.Strings("args", cmd.Args),
		zap.String("result", result),
		zap.Bool("error", IsError(result)),
		zap.String("session_id", sessionID))

	if oldUsr != session.User {
		err := c.userStorage.SaveRaw(ctx, session.User)
		logger.Debug(
			"save changed user failed",
			zap.Any("old_user", oldUsr),
			zap.Any("new_user", session.User),
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
	}

	return result

}

// del - executes the del command to remove a key from the storage
func (c *Database) del(ctx context.Context, user *models.User, args []string) string {
	if !user.ActiveRole.Del {
		return WrapError(ErrPermissionDenied)
	}

	key := storage.MakeKey(user.ActiveRole.Namespace, args[0])
	if err := c.storage.Del(ctx, key); err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// get - executes the get command to retrieve the value of a key from the storage.
func (c *Database) get(ctx context.Context, user *models.User, args []string) string {
	if !user.ActiveRole.Get {
		return WrapError(ErrPermissionDenied)
	}

	key := storage.MakeKey(user.ActiveRole.Namespace, args[0])
	val, err := c.storage.Get(ctx, key)
	if err != nil {
		return WrapError(err)
	}

	return WrapOK(val)
}

// set - executes the SET command to store a key-value pair in the storage.
func (c *Database) set(ctx context.Context, user *models.User, args []string) string {
	if !user.ActiveRole.Set {
		return WrapError(ErrPermissionDenied)
	}

	key := storage.MakeKey(user.ActiveRole.Namespace, args[0])
	if err := c.storage.Set(ctx, key, args[1]); err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// Login - authenticates a user based on the provided query.
func (c *Database) Login(ctx context.Context, sessionID string, query string) (*models.User, error) {
	cmd, err := c.parser.Parse(query)
	if err != nil {
		logger.Debug("parse query failed", zap.Error(err))
		return nil, fmt.Errorf("parse input failed: %w", err)
	}

	var user *models.User
	if cmd.Type == compute.CommandAUTH {
		user, err = c.userStorage.Authenticate(ctx, cmd.Args[0], cmd.Args[1])
		if err != nil {
			return nil, err
		}
	}

	if user == nil {
		return nil, ErrAuthenticationRequired
	}

	if err = c.sessions.Create(sessionID, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Logout - logs out the user by deleting their session token.
func (c *Database) Logout(ctx context.Context, sessionID string) string {
	c.sessions.Delete(sessionID)

	return okPrefix
}

// help - executes the help command to print information about commands.
func (c *Database) help(ctx context.Context, usr *models.User, _ []string) string {
	if usr.IsAdmin(c.cfg) {
		return compute.AdminHelpText
	}

	return compute.UserHelpText
}

// help - executes the help command to print information about commands.
func (c *Database) listSessions(ctx context.Context, _ *models.User, _ []string) string {
	sessions := c.sessions.List()
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
func (c *Database) ns(ctx context.Context, _ *models.User, _ []string) string {
	namepaces, err := c.namespaceStorage.List(ctx)
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
func (c *Database) createUser(ctx context.Context, _ *models.User, args []string) string {
	usr, err := c.userStorage.Create(ctx, args[0], args[1])
	if err != nil {
		return WrapError(err)
	}

	if _, err := c.userStorage.Append(ctx, usr.Username); err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// createUser - executes the create user command to create a new user.
func (c *Database) deleteUser(ctx context.Context, usr *models.User, args []string) string {
	if err := c.userStorage.Delete(ctx, args[0]); err != nil {
		return WrapError(err)
	}

	if _, err := c.userStorage.Remove(ctx, usr.Username); err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// assignRole - executes the assign role command to assign a role to a user.
func (c *Database) assignRole(ctx context.Context, _ *models.User, args []string) string {
	if err := c.userStorage.AssignRole(ctx, args[0], args[1]); err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// assignRole - executes the assign role command to assign a role to a user.
func (c *Database) divestRole(ctx context.Context, _ *models.User, args []string) string {
	if err := c.userStorage.AssignRole(ctx, args[0], args[1]); err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// users - executes the users command to list all usernames.
func (c *Database) users(ctx context.Context, _ *models.User, args []string) string {
	users, err := c.userStorage.ListUsernames(ctx)
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

func (c *Database) getUser(ctx context.Context, _ *models.User, args []string) string {
	user, err := c.userStorage.Get(ctx, args[0])
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

func (c *Database) getRole(ctx context.Context, _ *models.User, args []string) string {
	role, err := c.rolesStorage.Get(ctx, args[0])
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
func (c *Database) createRole(ctx context.Context, _ *models.User, args []string) string {
	if !c.namespaceStorage.Exists(ctx, args[2]) {
		return WrapError(identity.ErrNamespaceNotFound)
	}

	_, err := c.rolesStorage.Get(ctx, args[0])
	if err != nil && !errors.Is(err, identity.ErrRoleNotFound) {
		return WrapError(err)
	}

	role, err := models.NewRole(args[0], args[1], args[2])
	if err != nil {
		return WrapError(err)
	}

	if err := c.rolesStorage.Save(ctx, &role); err != nil {
		return WrapError(err)
	}

	if _, err := c.rolesStorage.Append(ctx, role.Name); err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// delRole - executes command to delete a role.
func (c *Database) delRole(ctx context.Context, _ *models.User, args []string) string {
	users, err := c.userStorage.ListUsernames(ctx)
	if err != nil && !errors.Is(err, identity.ErrEmptyUsers) {
		return WrapError(err)
	}

	for _, username := range users {
		user, err := c.userStorage.Get(ctx, username)
		if err != nil {
			return WrapError(err)
		}

		if slices.Contains(user.Roles, args[0]) {
			return WrapError(fmt.Errorf("cannot delete assigned to user '%s' role", username))
		}
	}

	if err := c.rolesStorage.Delete(ctx, args[0]); err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// listRoles - executes the listRoles command to list all listRoles.
func (c *Database) listRoles(ctx context.Context, _ *models.User, args []string) string {
	roles, err := c.rolesStorage.List(ctx)
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
func (c *Database) me(ctx context.Context, user *models.User, _ []string) string {
	return WrapOK(fmt.Sprintf(
		"user: '%s', roles: '%s', ns: '%s', perms: '%s'",
		user.Username, strings.Join(user.Roles, " "),
		user.ActiveRole.Namespace, user.ActiveRole.Perms(),
	))
}

// createNS - executes the create ns command to create a new namespace.
func (c *Database) createNS(ctx context.Context, _ *models.User, args []string) string {
	err := c.namespaceStorage.Save(ctx, args[0])
	if err != nil {
		return WrapError(err)
	}

	if _, err = c.namespaceStorage.Append(ctx, args[0]); err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// deleteNS - executes the delete ns command to delete a namespace
func (c *Database) deleteNS(ctx context.Context, _ *models.User, args []string) string {
	roles, err := c.rolesStorage.List(ctx)
	if err != nil {
		return WrapError(err)
	}

	for _, name := range roles {
		role, err := c.rolesStorage.Get(ctx, name)
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

	err = c.namespaceStorage.Delete(ctx, args[0])
	if err != nil {
		return WrapError(err)
	}

	return okPrefix
}

// setNamespace - executes the set ns command to change the current namespace for the user.
func (c *Database) setNamespace(ctx context.Context, user *models.User, args []string) string {
	namespace := args[0]

	if !c.namespaceStorage.Exists(ctx, namespace) {
		return WrapError(identity.ErrNamespaceNotFound)
	}

	var (
		role      *models.Role
		hasAccess bool
	)
	if !user.IsAdmin(c.cfg) {
		for _, roleName := range user.Roles {
			var err error
			role, err = c.rolesStorage.Get(ctx, roleName)
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

	if user.IsAdmin(c.cfg) {
		role = &models.Role{
			Get: true, Set: true, Del: true,
			Namespace: namespace,
		}
	}

	user.ActiveRole = *role
	return okPrefix
}
