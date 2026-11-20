package ownframe

import (
	"github.com/chinmay-sawant/ownframe/internal/clipboard"
	"github.com/chinmay-sawant/ownframe/internal/secret"
)

// ErrUnavailable reports a failed OS clipboard read outside memory-only
// mode. It carries no fallback text.
var ErrUnavailable = clipboard.ErrUnavailable

// UseMemory forces the in-memory clipboard. Tests call this before any
// Write so they never touch the desktop clipboard.
func UseMemory(on bool) { clipboard.UseMemory(on) }

// Write stores text in memory-only mode, and tries the OS clipboard
// otherwise.
func Write(text string) { clipboard.Write(text) }

// Read returns the memory text in memory-only mode, and the OS clipboard
// otherwise. An OS failure returns ErrUnavailable and no text.
func Read() (string, error) { return clipboard.Read() }

// Clear empties the clipboard. Tests call UseMemory(true) first so Clear
// never touches the desktop.
func Clear() { clipboard.Clear() }

// Secret wraps a password, token, or key. Format verbs, String, GoString,
// and slog all render [REDACTED]. Call Reveal to use it.
type Secret = secret.Secret
