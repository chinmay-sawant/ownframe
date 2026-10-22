//go:build linux && !android

package clipboard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWaylandClear(t *testing.T) {
	dir := withWaylandStub(t)
	writeOS("secret")
	clearOS()

	if got, ok := readOS(); ok || got != "" {
		t.Fatalf("read = %q %v", got, ok)
	}

	args, err := os.ReadFile(filepath.Join(dir, "copy-args"))
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

	args, err := os.ReadFile(filepath.Join(dir, "copy-args"))
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
