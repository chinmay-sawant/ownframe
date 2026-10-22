//go:build linux && !android

package clipboard

import (
	"os"
	"path/filepath"
	"testing"
)

func writeWLStub(t *testing.T, dir string) {
	t.Helper()
	writeTool(t, dir, "wl-copy", `#!/bin/sh
d=$(dirname "$0")
printf '%s\n' "$@" > "$d/copy-args"
if [ "$1" = "--clear" ]; then
  rm -f "$d/clip"
  exit 0
fi
cat > "$d/clip"
`)
	writeTool(t, dir, "wl-paste", `#!/bin/sh
d=$(dirname "$0")
printf '%s\n' "$@" > "$d/paste-args"
if [ ! -f "$d/clip" ]; then
  exit 1
fi
cat "$d/clip"
`)
}

func withWaylandStub(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeWLStub(t, dir)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("WAYLAND_DISPLAY", "wayland-test")
	t.Setenv("DISPLAY", "")

	return dir
}

func TestWaylandRoundTrip(t *testing.T) {
	dir := withWaylandStub(t)
	tests := []struct {
		in string
	}{
		{"hello"},
		{"a\nb"},
		{"--flag"},
		{""},
	}

	for _, tt := range tests {
		writeOS(tt.in)
		got, ok := readOS()
		if !ok || got != tt.in {
			t.Fatalf("read %q = %q %v", tt.in, got, ok)
		}
	}

	args, err := os.ReadFile(filepath.Join(dir, "copy-args"))
	if err != nil {
		t.Fatal(err)
	}
	if string(args) != "--type\ntext/plain\n--\n" {
		t.Fatalf("copy-args %q", args)
	}

	paste, err := os.ReadFile(filepath.Join(dir, "paste-args"))
	if err != nil {
		t.Fatal(err)
	}
	if string(paste) != "--no-newline\n" {
		t.Fatalf("paste-args %q", paste)
	}
}
