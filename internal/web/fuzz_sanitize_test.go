package web

import (
	"strings"
	"testing"
)

// FuzzSanitizeLine feeds hostile log lines to the sanitizer (GEN-24).
// Output never carries CR/LF and stays within the length cap.
func FuzzSanitizeLine(f *testing.F) {
	for _, seed := range []string{"ok", "a\nb\rc", "secret /tmp/x", strings.Repeat("a", 600), "", "\x00\x1b[2J"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, s string) {
		out := sanitizeLine(s)
		if strings.ContainsAny(out, "\r\n") {
			t.Fatal("unsanitized line break")
		}
		if len(out) > sanitizeMaxLen {
			t.Fatal("over cap")
		}
	})
}
