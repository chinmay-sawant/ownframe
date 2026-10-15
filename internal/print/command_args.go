package print

import (
	"errors"
	"strings"
	"unicode"
)

// ErrBadPath means the print path contains a newline or control character
// and was rejected before any command was invoked.
var ErrBadPath = errors.New("ownframe: print path contains a control character")

// badPath reports whether path contains a newline or other control character.
func badPath(path string) bool {
	for _, r := range path {
		if r == '\n' || r == '\r' || unicode.IsControl(r) {
			return true
		}
	}

	return false
}

func validPath(path string) error {
	if badPath(path) {
		return ErrBadPath
	}

	return nil
}

func lpFileArg(path string) []string {
	return []string{"--", path}
}

func openFileArg(path string) []string {
	return []string{"--", path}
}

func appleScriptPrint(path string) []string {
	return []string{"-e", `on run argv`, "-e", `tell application "Preview" to print (POSIX file (item 1 of argv))`, "-e", `end run`, "--", path}
}

func powershellPrint(path string) []string {
	quoted := strings.ReplaceAll(path, "'", "''")

	return []string{
		"-NoProfile",
		"-Command",
		"Start-Process -FilePath '" + quoted + "' -Verb Print",
	}
}
