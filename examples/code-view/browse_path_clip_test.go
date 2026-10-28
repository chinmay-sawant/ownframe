package main

import (
	"strings"
	"testing"
)

func TestClipBrowsePath(t *testing.T) {
	if got := clipBrowsePath(1280, "projects"); got != "projects" {
		t.Fatalf("short name clipped to %q", got)
	}

	// budget = (560 - 22 - 44 - 6 - 8) / 6.5 = 73; the 80-rune name keeps
	// 72 runes plus the ellipsis.
	name := strings.Repeat("0123456789", 8)
	if got, want := len([]rune(name)), 80; got != want {
		t.Fatalf("fixture name is %d runes, want %d", got, want)
	}

	got := clipBrowsePath(1280, name)
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("long name %q has no ellipsis", got)
	}

	if got, want := len([]rune(got)), 73; got != want {
		t.Fatalf("clipped name is %d runes, want %d", got, want)
	}

	// window 400 px: panel = 400 - 40 = 360, budget = (360 - 80) / 6.5 = 43.
	narrow := clipBrowsePath(400, name)
	if got, want := len([]rune(narrow)), 43; got != want {
		t.Fatalf("narrow clipped name is %d runes, want %d", got, want)
	}

	if got := clipBrowsePath(1280, ""); got != "" {
		t.Fatalf("empty name = %q, want empty", got)
	}

	if got := clipBrowsePath(100, name); got != "…" {
		t.Fatalf("tiny window clipped to %q, want ellipsis only", got)
	}
}
