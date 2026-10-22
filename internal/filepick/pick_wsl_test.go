//go:build linux && !android

package filepick

import (
	"context"
	"errors"
	"os"
	"testing"
)

const (
	powerShellPick = "#!/bin/sh\nprintf '%s\\n' 'C:\\Users\\me\\notes.txt'\n"
	wslpathConvert = "#!/bin/sh\n[ \"$1\" = \"-u\" ] || exit 2\nprintf '%s\\n' /mnt/c/Users/me/notes.txt\n"
)

func TestInWSL(t *testing.T) {
	t.Setenv("WSL_DISTRO_NAME", "Ubuntu")

	if !inWSL() {
		t.Fatal("inWSL = false with WSL_DISTRO_NAME")
	}
}

func TestPickWindows(t *testing.T) {
	dir := t.TempDir()
	writeTool(t, dir, "powershell.exe", powerShellPick)
	writeTool(t, dir, "wslpath", wslpathConvert)
	t.Setenv("PATH", dir)

	got, err := pickWindows(context.Background(), "Open")
	if err != nil || got != "/mnt/c/Users/me/notes.txt" {
		t.Fatalf("pick = %q %v", got, err)
	}
}

func TestPickWindowsCancel(t *testing.T) {
	dir := t.TempDir()
	writeTool(t, dir, "powershell.exe", "#!/bin/sh\nexit 0\n")

	got, err := pickWindowsAt(context.Background(), "Open",
		[]string{dir + "/powershell.exe"})
	if !errors.Is(err, ErrCanceled) || got != "" {
		t.Fatalf("pick = %q %v", got, err)
	}
}

func TestPickWindowsShellError(t *testing.T) {
	dir := t.TempDir()
	writeTool(t, dir, "powershell.exe", "#!/bin/sh\nexit 1\n")

	_, err := pickWindowsAt(context.Background(), "Open",
		[]string{dir + "/powershell.exe"})
	if !errors.Is(err, ErrNoDialog) {
		t.Fatalf("err = %v", err)
	}
}

func TestPickWindowsNoShell(t *testing.T) {
	if _, err := pickWindowsAt(context.Background(), "Open", nil); !errors.Is(err, ErrNoDialog) {
		t.Fatal(err)
	}
}

func TestShellPathsFindsStandardPowerShell(t *testing.T) {
	const std = "/mnt/c/Windows/System32/WindowsPowerShell/v1.0/powershell.exe"

	if _, err := os.Stat(std); err != nil {
		t.Skip("no standard PowerShell install")
	}

	t.Setenv("PATH", t.TempDir())

	for _, path := range shellPaths() {
		if path == std {
			return
		}
	}

	t.Fatalf("shellPaths = %v, want %s", shellPaths(), std)
}
