package models

import "errors"

var (
	ErrNamespaceAlreadyExists = errors.New("namespace already exists")
	ErrNamespaceNotFound      = errors.New("namespace not found")
	ErrEmptyNamespaces        = errors.New("empty namespaces")

	ErrRoleAlreadyExists = errors.New("role already exists")
	ErrRoleNotFound      = errors.New("role not found")
	ErrEmptyRoles        = errors.New("empty roles")
	ErrInvalidPerms      = errors.New("invalid perms: perms must contain only 'r', 'w', 'd'")

	ErrAuthenticationFailed   = errors.New("authentication failed")
	ErrAuthenticationRequired = errors.New("authentication required")
	ErrPermissionDenied       = errors.New("permission denied")
	ErrUserAlreadyExists      = errors.New("user already exists")
	ErrUserNotFound           = errors.New("user not found")
	ErrEmptyUsers             = errors.New("empty users")

	ErrKeyNotFound      = errors.New("key not found")
	ErrInvalidOperation = errors.New("invalid operation")
)
