package database

import (
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
)

// WrapError - wrapping error with prefix '[error]'.
func WrapError(err error) string {
	return fmt.Sprintf("[error] %s", err.Error())
}

// WrapError - wrapping message with prefix '[ok]'.
func WrapOK(msg string) string {
	if msg == "" {
		return "[ok]"
	}

	return "[ok] " + msg
}

// isError - check the prefix 'error:' exists.
func isError(val string) bool {
	return strings.Contains(val, "[error]")
}

// HandleQuery -  processes a user query by parsing and executing the corresponding command.
func (c *Database) HandleQuery(user *models.User, query string) string {
	cmd, err := c.parser.Parse(query)
	if err != nil {
		logger.Debug("parse query failed", zap.Error(err))
		return WrapError(fmt.Errorf("parse input failed: %w", err))
	}

	logger.Info("parsed command",
		zap.Stringer("cmd_type", cmd.Type),
		zap.Strings("args", cmd.Args))

	handler, ok := map[compute.CommandType]CommandHandler{
		compute.CommandCREATEUSER:      {Func: c.createUser, AdminOnly: true},
		compute.CommandASSIGNROLE:      {Func: c.assignRole, AdminOnly: true},
		compute.CommandCREATEROLE:      {Func: c.createRole, AdminOnly: true},
		compute.CommandDELETEROLE:      {Func: c.Delete, AdminOnly: true},
		compute.CommandROLES:           {Func: c.roles, AdminOnly: true},
		compute.CommandUSERS:           {Func: c.users, AdminOnly: true},
		compute.CommandCREATENAMESPACE: {Func: c.createNS, AdminOnly: true},
		compute.CommandDELETENAMESPACE: {Func: c.deleteNS, AdminOnly: true},
		compute.CommandNAMESPACES:      {Func: c.ns, AdminOnly: true},
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

	if handler.AdminOnly && user.Username != c.cfg.Username {
		return WrapError(ErrPermissionDenied)
	}

	oldUsr := user
	result := handler.Func(user, cmd.Args)
	logger.Info("operation executed",
		zap.Stringer("cmd_type", cmd.Type),
		zap.Strings("args", cmd.Args),
		zap.String("result", result),
		zap.Bool("error", isError(result)))

	if oldUsr != user {
		err := c.userStorage.SaveRaw(user)
		logger.Debug(
			"save changed user failed",
			zap.Any("old_user", oldUsr),
			zap.Any("new_user", user),
			zap.Error(err),
		)
	}

	return result

}

// Login - authenticates a user based on the provided query.
func (c *Database) Login(query string) (*models.User, error) {
	cmd, err := c.parser.Parse(query)
	if err != nil {
		logger.Debug("parse query failed", zap.Error(err))
		return nil, fmt.Errorf("parse input failed: %w", err)
	}

	var user *models.User
	if cmd.Type == compute.CommandAUTH {
		user, err = c.userStorage.Authenticate(cmd.Args[0], cmd.Args[1])
		if err != nil {
			return nil, err
		}
	}

	if user == nil {
		return nil, ErrAuthenticationRequired
	}

	token, err := c.sessions.Create(user.Username)
	if err != nil {
		return nil, err
	}
	user.Token = token

	return user, nil
}

// Logout - logs out the user by deleting their session token.
func (c *Database) Logout(user *models.User, args []string) string {
	err := c.sessions.Delete(user.Token)
	if err != nil {
		return WrapError(err)
	}

	return WrapOK("")
}

// del - executes the del command to remove a key from the storage
func (c *Database) del(user *models.User, args []string) string {
	if !user.Cur.Del {
		return WrapError(ErrPermissionDenied)
	}

	key := storage.MakeKey(user.Cur.Namespace, args[0])
	if err := c.storage.Del(key); err != nil {
		return WrapError(err)
	}

	return WrapOK(args[0])
}

// get - executes the get command to retrieve the value of a key from the storage.
func (c *Database) get(user *models.User, args []string) string {
	if !user.Cur.Get {
		return WrapError(ErrPermissionDenied)
	}

	key := storage.MakeKey(user.Cur.Namespace, args[0])
	val, err := c.storage.Get(key)
	if err != nil {
		return WrapError(err)
	}

	return WrapOK(val)
}

// set - executes the SET command to store a key-value pair in the storage.
func (c *Database) set(user *models.User, args []string) string {
	if !user.Cur.Set {
		return WrapError(ErrPermissionDenied)
	}

	key := storage.MakeKey(user.Cur.Namespace, args[0])
	if err := c.storage.Set(key, args[1]); err != nil {
		return WrapError(err)
	}

	return WrapOK(args[1])
}

// help - executes the help command to print information about commands.
func (c *Database) help(usr *models.User, _ []string) string {
	if usr.IsAdmin(c.cfg) {
		return compute.AdminHelpText
	}

	return compute.UserHelpText
}

// ns - executes the ns command to list namespaces.
func (c *Database) ns(_ *models.User, _ []string) string {
	namepaces, err := c.namespaceStorage.List()
	if err != nil {
		return WrapError(err)
	}

	return WrapOK(strings.Join(namepaces, " "))
}

// createUser - executes the create user command to create a new user.
func (c *Database) createUser(user *models.User, args []string) string {
	usr, err := c.userStorage.Create(args[0], args[1])
	if err != nil {
		return WrapError(err)
	}

	if _, err := c.userStorage.Append(usr.Username); err != nil {
		return WrapError(err)
	}

	return WrapOK(user.Username)
}

// assignRole - executes the assign role command to assign a role to a user.
func (c *Database) assignRole(user *models.User, args []string) string {
	role, err := c.rolesStorage.Get(args[1])
	if err != nil {
		return WrapError(err)
	}

	if err := c.userStorage.AssignRole(args[0], role.Name); err != nil {
		return WrapError(err)
	}

	return WrapOK("")
}

// users - executes the users command to list all usernames.
func (c *Database) users(user *models.User, args []string) string {
	list, err := c.userStorage.ListUsernames()
	if err != nil {
		return WrapError(err)
	}

	return WrapOK(strings.Join(list, " "))
}

// createRole - executes the create role command to create a new role.
func (c *Database) createRole(user *models.User, args []string) string {
	if !c.namespaceStorage.Exists(args[2]) {
		return WrapError(identity.ErrNamespaceNotFound)
	}

	_, err := c.rolesStorage.Get(args[0])
	if err != nil && !errors.Is(err, identity.ErrRoleNotFound) {
		return WrapError(err)
	}

	role, err := models.NewRole(args[0], args[1], args[2])
	if err != nil {
		return WrapError(err)
	}

	if err := c.rolesStorage.Save(&role); err != nil {
		return WrapError(err)
	}

	if _, err := c.rolesStorage.Append(role.Name); err != nil {
		return WrapError(err)
	}

	return WrapOK("")
}

// Delete - executes command to delete a role.
func (c *Database) Delete(user *models.User, args []string) string {
	users, err := c.userStorage.ListUsernames()
	if err != nil && !errors.Is(err, identity.ErrEmptyUsers) {
		return WrapError(err)
	}

	for _, username := range users {
		user, err := c.userStorage.Get(username)
		if err != nil {
			return WrapError(err)
		}

		if slices.Contains(user.Roles, args[0]) {
			return WrapError(fmt.Errorf("cannot delete assigned to user '%s' role", username))
		}
	}

	if err := c.rolesStorage.Delete(args[0]); err != nil {
		return WrapError(err)
	}

	return WrapOK("")
}

// roles - executes the roles command to list all roles.
func (c *Database) roles(user *models.User, args []string) string {
	roles, err := c.rolesStorage.List()
	if err != nil {
		return WrapError(err)
	}

	return WrapOK(strings.Join(roles, ", "))
}

// me - executes the me command to display information about the current user,
// including their username, roles, namespace, and permissions.
func (c *Database) me(user *models.User, _ []string) string {
	return WrapOK(fmt.Sprintf(
		"user: '%s', roles: '%s', ns: '%s', perms: '%s'",
		user.Username, strings.Join(user.Roles, " "),
		user.Cur.Namespace, user.Cur.Perms(),
	))
}

// createNS - executes the create ns command to create a new namespace.
func (c *Database) createNS(user *models.User, args []string) string {
	err := c.namespaceStorage.Save(args[0])
	if err != nil {
		return WrapError(err)
	}

	if _, err = c.namespaceStorage.Append(args[0]); err != nil {
		return WrapError(err)
	}

	return WrapOK("")
}

// deleteNS - executes the delete ns command to delete a namespace
func (c *Database) deleteNS(user *models.User, args []string) string {
	err := c.namespaceStorage.Delete(args[0])
	if err != nil {
		return WrapError(err)
	}

	return WrapOK("")
}

// setNamespace - executes the set ns command to change the current namespace for the user.
func (c *Database) setNamespace(user *models.User, args []string) string {
	namespace := args[0]

	if !c.namespaceStorage.Exists(namespace) {
		return WrapError(identity.ErrNamespaceNotFound)
	}

	var (
		role      *models.Role
		hasAccess bool
	)
	if !user.IsAdmin(c.cfg) {
		for _, roleName := range user.Roles {
			var err error
			role, err = c.rolesStorage.Get(roleName)
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

	user.Cur = *role
	return WrapOK("")
}
