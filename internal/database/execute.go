package database

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/neekrasov/kvdb/internal/database/command"
	"github.com/neekrasov/kvdb/internal/database/storage"
	"github.com/neekrasov/kvdb/internal/database/storage/models"
	"github.com/neekrasov/kvdb/pkg/logger"
	"go.uber.org/zap"
)

// wrapError - wrapping error with prefix 'error:'.
func wrapError(err error) string {
	return fmt.Sprintf("error: %s", err.Error())
}

// isError - check the prefix 'error:' exists.
func isError(val string) bool {
	return strings.Contains(val, "error:")
}

// HandleQuery processes a user query by parsing and executing the corresponding command.
func (c *Database) HandleQuery(user *models.User, query string) string {
	cmd, err := c.parser.Parse(query)
	if err != nil {
		logger.Debug("parse query failed", zap.Error(err))
		return wrapError(fmt.Errorf("parse input failed: %w", err))
	}

	logger.Info("parsed command",
		zap.Stringer("cmd_type", cmd.Type),
		zap.Strings("args", cmd.Args))

	handler, ok := map[command.CommandType]CommandHandler{
		command.CommandCREATEUSER:      {Func: c.createUser, AdminOnly: true},
		command.CommandASSIGNROLE:      {Func: c.assignRole, AdminOnly: true},
		command.CommandCREATEROLE:      {Func: c.createRole, AdminOnly: true},
		command.CommandDELETEROLE:      {Func: c.Delete, AdminOnly: true},
		command.CommandROLES:           {Func: c.roles, AdminOnly: true},
		command.CommandUSERS:           {Func: c.users, AdminOnly: true},
		command.CommandCREATENAMESPACE: {Func: c.createNS, AdminOnly: true},
		command.CommandDELETENAMESPACE: {Func: c.deleteNS, AdminOnly: true},
		command.CommandSETNS:           {Func: c.setNamespace},
		command.CommandME:              {Func: c.me},
		command.CommandGET:             {Func: c.get},
		command.CommandSET:             {Func: c.set},
		command.CommandDEL:             {Func: c.del},
	}[cmd.Type]

	if !ok {
		return wrapError(models.ErrInvalidOperation)
	}

	if handler.AdminOnly && user.Username != c.cfg.Username {
		return wrapError(models.ErrPermissionDenied)
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

// Login - Authenticates a user based on the provided query.
func (c *Database) Login(query string) (*models.User, error) {
	cmd, err := c.parser.Parse(query)
	if err != nil {
		logger.Debug("parse query failed", zap.Error(err))
		return nil, fmt.Errorf("parse input failed: %w", err)
	}

	logger.Info("parsed command",
		zap.Stringer("cmd_type", cmd.Type),
		zap.Strings("args", cmd.Args))

	var user *models.User
	if cmd.Type == command.CommandAUTH {
		user, err = c.userStorage.Authenticate(cmd.Args[0], cmd.Args[1])
		if err != nil {
			return nil, err
		}
	}

	if user == nil {
		return nil, models.ErrAuthenticationRequired
	}

	token, err := c.sessions.Create(user.Username)
	if err != nil {
		return nil, err
	}
	user.Token = token

	return user, nil
}

// Logout - Logs out the user by deleting their session token.
func (c *Database) Logout(user *models.User, args []string) string {
	err := c.sessions.Delete(user.Token)
	if err != nil {
		return wrapError(err)
	}

	return "OK"
}

// del - Executes the DEL command to remove a key from the storage
func (c *Database) del(user *models.User, args []string) string {
	if !user.Cur.Del {
		return wrapError(models.ErrPermissionDenied)
	}

	key := storage.MakeKey(user.Cur.Namespace, args[0])
	if err := c.storage.Del(key); err != nil {
		return wrapError(err)
	}

	return args[0]
}

// get - Executes the GET command to retrieve the value of a key from the storage.
func (c *Database) get(user *models.User, args []string) string {
	if !user.Cur.Get {
		return wrapError(models.ErrPermissionDenied)
	}

	key := storage.MakeKey(user.Cur.Namespace, args[0])
	val, err := c.storage.Get(key)
	if err != nil {
		return wrapError(err)
	}

	return val
}

// set - Executes the SET command to store a key-value pair in the storage.
func (c *Database) set(user *models.User, args []string) string {
	if !user.Cur.Set {
		return wrapError(models.ErrPermissionDenied)
	}

	key := storage.MakeKey(user.Cur.Namespace, args[0])
	c.storage.Set(key, args[1])

	return args[1]
}

// createUser - Executes the CREATEUSER command to create a new user.
func (c *Database) createUser(user *models.User, args []string) string {
	usr, err := c.userStorage.Create(args[0], args[1])
	if err != nil {
		return wrapError(err)
	}

	if _, err := c.userStorage.Append(usr.Username); err != nil {
		return wrapError(err)
	}

	return usr.Username
}

// assignRole - Executes the ASSIGNROLE command to assign a role to a user.
func (c *Database) assignRole(user *models.User, args []string) string {
	role, err := c.rolesStorage.Get(args[1])
	if err != nil {
		return wrapError(err)
	}

	if err := c.userStorage.AssignRole(args[0], role.Name); err != nil {
		return wrapError(err)
	}

	return "OK"
}

// users - Executes the USERS command to list all usernames.
func (c *Database) users(user *models.User, args []string) string {
	list, err := c.userStorage.ListUsernames()
	if err != nil {
		return wrapError(err)
	}

	return strings.Join(list, " ")
}

// createRole - Executes the CREATEROLE command to create a new role
func (c *Database) createRole(user *models.User, args []string) string {
	if !c.namespaceStorage.Exists(args[2]) {
		return wrapError(models.ErrNamespaceNotFound)
	}

	_, err := c.rolesStorage.Get(args[0])
	if err != nil && !errors.Is(err, models.ErrRoleNotFound) {
		return wrapError(err)
	}

	role, err := models.NewRole(args[0], args[1], args[2])
	if err != nil {
		return wrapError(err)
	}

	if err := c.rolesStorage.Save(&role); err != nil {
		return wrapError(err)
	}

	if _, err := c.rolesStorage.Append(role.Name); err != nil {
		return wrapError(err)
	}

	return "OK"
}

// Delete - Executes the Delete command to delete a role.
func (c *Database) Delete(user *models.User, args []string) string {
	users, err := c.userStorage.ListUsernames()
	if err != nil && !errors.Is(err, models.ErrEmptyUsers) {
		return wrapError(err)
	}

	for _, username := range users {
		user, err := c.userStorage.Get(username)
		if err != nil {
			return wrapError(err)
		}

		if slices.Contains(user.Roles, args[0]) {
			return wrapError(fmt.Errorf("cannot delete assigned to user '%s' role", username))
		}
	}

	if err := c.rolesStorage.Delete(args[0]); err != nil {
		return wrapError(err)
	}

	return "OK"
}

// roles - Executes the ROLES command to list all roles.
func (c *Database) roles(user *models.User, args []string) string {
	roles, err := c.rolesStorage.List()
	if err != nil {
		return wrapError(err)
	}

	return strings.Join(roles, ", ")
}

// me - Executes the ME command to display information about the current user,
// including their username, roles, namespace, and permissions.
func (c *Database) me(user *models.User, _ []string) string {
	fmt.Println(user.Cur)
	return fmt.Sprintf(
		"user: '%s', roles: '%s', ns: '%s', perms: '%s'",
		user.Username, strings.Join(user.Roles, " "),
		user.Cur.Namespace, user.Cur.Perms(),
	)
}

// createNS - Executes the CREATENAMESPACE command to create a new namespace.
func (c *Database) createNS(user *models.User, args []string) string {
	err := c.namespaceStorage.Save(args[0])
	if err != nil {
		return wrapError(err)
	}

	if _, err = c.namespaceStorage.Append(args[0]); err != nil {
		return wrapError(err)
	}

	return "OK"
}

// deleteNS - Executes the DELETENAMESPACE command to delete a namespace
func (c *Database) deleteNS(user *models.User, args []string) string {
	err := c.namespaceStorage.Delete(args[0])
	if err != nil {
		return wrapError(err)
	}

	return "OK"
}

// setNamespace - Executes the SETNS command to change the current namespace for the user.
func (c *Database) setNamespace(user *models.User, args []string) string {
	namespace := args[0]

	if !c.namespaceStorage.Exists(namespace) {
		return wrapError(models.ErrNamespaceNotFound)
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
			return wrapError(models.ErrPermissionDenied)
		}
	}

	if user.IsAdmin(c.cfg) {
		hasAccess = true
		role = &models.Role{
			Get: true, Set: true, Del: true,
			Namespace: namespace,
		}
	}

	user.Cur = *role
	return "OK"
}
