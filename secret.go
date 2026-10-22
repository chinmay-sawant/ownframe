package ownframe

import "github.com/chinmay-sawant/ownframe/internal/secret"

// Secret wraps a password, token, or key. Format verbs, String, GoString,
// and slog all render [REDACTED]. Call Reveal to use it.
type Secret = secret.Secret
