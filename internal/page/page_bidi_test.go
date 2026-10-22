package page

import (
	"strings"
	"testing"
)

func TestSanitizeBidiStrips(t *testing.T) {
	t.Parallel()
	got := SanitizeBidi("a\u202eb\u2066c")
	if got != "abc" {
		t.Fatalf("got %q", got)
	}
	if err := CheckIdentifier("alice_01"); err != nil {
		t.Fatalf("valid: %v", err)
	}
	if err := CheckIdentifier("a\u202eb"); err == nil {
		t.Fatal("bidi identifier allowed")
	}
	if !strings.Contains(SanitizeBidi("x"), "x") {
		t.Fatal("plain text changed")
	}
}
