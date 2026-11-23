// Package crash writes a panic report to a text file on this machine.
// The file is not uploaded.
package crash

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

var (
	mu     sync.Mutex
	custom string
)

// crashDirPerm keeps crash reports private to the owner.
const crashDirPerm = 0o700

// Write creates the folder and writes one UTF-8 report.
// The name starts with a UTC timestamp, and a later call in that
// same second gets a different name.
//
// Newlines in title and reason become spaces, so neither can forge extra
// header lines in the report.
func Write(title, reason string) (string, error) {
	folder, shared := locate()

	_, statErr := os.Stat(folder)
	if err := os.MkdirAll(folder, crashDirPerm); err != nil {
		return "", err
	}
	if errors.Is(statErr, os.ErrNotExist) {
		if err := os.Chmod(folder, crashDirPerm); err != nil {
			return "", err
		}
	}
	if shared {
		if err := secureSharedDir(folder); err != nil {
			return "", err
		}
	}

	body := fmt.Sprintf(
		"title: %s\nreason: %s\ngo: %s\n%s",
		oneLine(title),
		oneLine(reason),
		runtime.Version(),
		debug.Stack(),
	)

	return create(folder, body)
}

// Dir returns the folder Write uses. It is always absolute.
func Dir() string {
	return location()
}

func location() string {
	dir, _ := locate()

	return dir
}

// locate returns the report folder and whether it sits in the shared temp
// directory, where another local user could have pre-created it.
func locate() (string, bool) {
	mu.Lock()
	dir := custom
	mu.Unlock()

	if dir != "" {
		return dir, false
	}
	config, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "ownframe-crashes-"+tempOwner()), true
	}

	return filepath.Join(config, "ownframe", "crashes"), false
}

// tempOwner names the per-user temp folder suffix. Windows reports a uid of
// -1, so it falls back to the account name, reduced to letters and digits.
func tempOwner() string {
	if uid := os.Getuid(); uid >= 0 {
		return strconv.Itoa(uid)
	}
	if u, err := user.Current(); err == nil {
		if name := alnum(u.Username); name != "" {
			return name
		}
	}

	return "user"
}

// alnum keeps ASCII letters and digits only.
func alnum(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r < unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r)) {
			b.WriteRune(r)
		}
	}

	return b.String()
}

// secureSharedDir refuses a symlinked folder in the shared temp directory and
// tightens a loose mode. Chmod fails on a folder another user owns, which
// rejects it.
func secureSharedDir(folder string) error {
	info, err := os.Lstat(folder)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("crash: %s is not a plain directory", folder)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != crashDirPerm {
		return os.Chmod(folder, crashDirPerm)
	}

	return nil
}

// oneLine turns every line break into a space.
func oneLine(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '\n', '\r', '\v', '\f', '\u0085', '\u2028', '\u2029':
			return ' '
		}

		return r
	}, s)
}
