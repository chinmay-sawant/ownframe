//go:build linux && !android

package clipboard

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// readWLFixture confines reads to the fixture directory owned by this test.
func readWLFixture(dir, name string) ([]byte, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	data, readErr := root.ReadFile(name)
	return data, errors.Join(readErr, root.Close())
}

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

	args, err := readWLFixture(dir, "copy-args")
	if err != nil {
		t.Fatal(err)
	}
	if string(args) != "--type\ntext/plain\n--\n" {
		t.Fatalf("copy-args %q", args)
	}

	paste, err := readWLFixture(dir, "paste-args")
	if err != nil {
		t.Fatal(err)
	}
	if string(paste) != "--no-newline\n" {
		t.Fatalf("paste-args %q", paste)
	}
}

func TestWaylandClear(t *testing.T) {
	dir := withWaylandStub(t)
	writeOS("secret")
	clearOS()

	if got, ok := readOS(); ok || got != "" {
		t.Fatalf("read = %q %v", got, ok)
	}

	args, err := readWLFixture(dir, "copy-args")
	if err != nil {
		t.Fatal(err)
	}
	if string(args) != "--clear\n" {
		t.Fatalf("copy-args %q", args)
	}
}

func TestWaylandWriteStaysOnStdin(t *testing.T) {
	dir := withWaylandStub(t)
	text := "evil; rm -rf / --help"
	writeOS(text)

	args, err := readWLFixture(dir, "copy-args")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(args), "evil") {
		t.Fatalf("user text in argv")
	}

	got, ok := readOS()
	if !ok || got != text {
		t.Fatalf("read = %q %v", got, ok)
	}
}

func writeTool(t *testing.T, dir, name, body string) {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o755); err != nil { //nolint:gosec // reason: stub wl-copy/wl-paste must be executable.
		t.Fatal(err)
	}
}

func TestWLArgs(t *testing.T) {
	tests := []struct {
		name string
		got  []string
		want []string
	}{
		{"copy", wlCopyArgs(), []string{"--type", "text/plain", "--"}},
		{"clear", wlClearArgs(), []string{"--clear"}},
		{"paste", wlPasteArgs(), []string{"--no-newline"}},
	}

	for _, tt := range tests {
		if !slices.Equal(tt.got, tt.want) {
			t.Fatalf("%s = %q", tt.name, tt.got)
		}
	}
}

func TestWaylandMissingTools(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	t.Setenv("WAYLAND_DISPLAY", "wayland-test")
	t.Setenv("DISPLAY", "")

	writeOS("dropped")
	if got, ok := readOS(); ok || got != "" {
		t.Fatalf("read = %q %v", got, ok)
	}

	clearOS()
}
