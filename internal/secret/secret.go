// Package secret holds credentials that must never print or log.
package secret

import (
	"encoding"
	"encoding/json"
	"fmt"
	"log/slog"
)

// redacted is the only rendering of a Secret outside Reveal.
const redacted = "[REDACTED]"

// Secret wraps a password, token, or key. Format verbs, String,
// GoString, and slog all render [REDACTED]. Call Reveal to use it.
type Secret string

var (
	_ json.Marshaler         = Secret("")
	_ encoding.TextMarshaler = Secret("")
)

// Reveal returns the wrapped value for actual use.
func (s Secret) Reveal() string { return string(s) }

// String hides the value.
func (s Secret) String() string { return redacted }

// GoString hides the value in %#v.
func (s Secret) GoString() string { return `secret.Secret("[REDACTED]")` }

// Format hides the value for every verb.
func (s Secret) Format(f fmt.State, _ rune) { _, _ = f.Write([]byte(redacted)) }

// LogValue hides the value in slog records.
func (s Secret) LogValue() slog.Value { return slog.StringValue(redacted) }

// MarshalJSON hides the value in JSON output.
func (s Secret) MarshalJSON() ([]byte, error) { return []byte(`"` + redacted + `"`), nil }

// MarshalText hides the value in text output.
func (s Secret) MarshalText() ([]byte, error) { return []byte(redacted), nil }
