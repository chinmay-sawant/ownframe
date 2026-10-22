package store

import "strings"

// isMissingTable reports SQLite's missing-table error.
func isMissingTable(err error) bool {
	return err != nil && strings.Contains(err.Error(), "no such table")
}
