package database

import (
	"fmt"
	"strings"
)

const (
	errPrefix = "[error]"
	okPrefix  = "[ok]"
)

// WrapError - wrapping error with prefix '[error]'.
func WrapError(err error) string {
	return fmt.Sprintf("%s %v", errPrefix, err)
}

// WrapError - wrapping message with prefix '[ok]'.
func WrapOK(msg string) string {
	if msg == "" {
		return okPrefix
	}

	return fmt.Sprintf("%s %s", okPrefix, msg)
}

// IsError - check the prefix 'error' exists.
func IsError(val string) bool {
	return strings.Contains(val, errPrefix)
}

// CutError - cat prefix 'error'.
func CutError(val string) (string, bool) {
	return strings.CutPrefix(val, errPrefix)
}

// CutOK - cat prefix 'ok'.
func CutOK(val string) (string, bool) {
	return strings.CutPrefix(val, okPrefix)
}
