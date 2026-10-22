//go:build linux && !android

package clipboard

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

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
