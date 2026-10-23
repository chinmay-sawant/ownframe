//go:build linux && !android

package filepick

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func writeTool(t *testing.T, dir, name, body string) {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o755); err != nil { //nolint:gosec // reason: the fake dialog tool must be executable to run.
		t.Fatal(err)
	}
}

func TestPickZenity(t *testing.T) {
	dir := t.TempDir()
	writeTool(t, dir, "zenity", `#!/bin/sh
[ "$1" = "--file-selection" ] || exit 2
echo /tmp/report.pdf
`)
	t.Setenv("PATH", dir)

	got, err := pickLinux(context.Background(), "Open")
	if err != nil || got != "/tmp/report.pdf" {
		t.Fatalf("pick = %q %v", got, err)
	}
}

func TestPickZenityCancel(t *testing.T) {
	dir := t.TempDir()
	writeTool(t, dir, "zenity", "#!/bin/sh\nexit 1\n")
	t.Setenv("PATH", dir)

	if got, err := pickLinux(context.Background(), "Open"); !errors.Is(err, ErrCanceled) {
		t.Fatalf("pick = %q %v", got, err)
	}
}

func TestPickKdialog(t *testing.T) {
	dir := t.TempDir()
	writeTool(t, dir, "kdialog", `#!/bin/sh
[ "$1" = "--getopenfilename" ] || exit 2
echo /tmp/kde.txt
`)
	t.Setenv("PATH", dir)

	got, err := pickLinux(context.Background(), "Open")
	if err != nil || got != "/tmp/kde.txt" {
		t.Fatalf("pick = %q %v", got, err)
	}
}

func TestPickZenityFirst(t *testing.T) {
	dir := t.TempDir()
	writeTool(t, dir, "zenity", "#!/bin/sh\necho /tmp/zen\n")
	writeTool(t, dir, "kdialog", "#!/bin/sh\necho /tmp/kde\n")
	t.Setenv("PATH", dir)

	got, err := pickLinux(context.Background(), "Open")
	if err != nil || got != "/tmp/zen" {
		t.Fatalf("pick = %q %v", got, err)
	}
}

func TestPickNoTool(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	if got, err := pickLinux(context.Background(), "Open"); !errors.Is(err, ErrNoDialog) {
		t.Fatalf("pick = %q %v", got, err)
	}
}
