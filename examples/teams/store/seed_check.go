package store

import (
	"errors"
	"strings"
)

// errMultiStatement means a seed line held stacked statements.
var errMultiStatement = errors.New("store: seed line holds stacked statements")

// singleStatement strips one trailing semicolon and rejects a line that
// still holds one, so Exec gets exactly one statement.
func singleStatement(line string) (string, bool) {
	stmt := strings.TrimSuffix(line, ";")
	if strings.TrimSpace(stmt) == "" || strings.Contains(stmt, ";") {
		return "", false
	}

	return stmt, true
}
