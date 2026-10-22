//go:build linux && !android

package filepick

import (
	"context"
	"testing"
)

func TestUncToWSL(t *testing.T) {
	cases := map[string]string{
		`\\wsl.localhost\Ubuntu\home\me\a.txt`: "/home/me/a.txt",
		`\\wsl$\Ubuntu\home\me\a.txt`:          "/home/me/a.txt",
	}

	for picked, want := range cases {
		got, ok := uncToWSL(picked)
		if !ok || got != want {
			t.Fatalf("uncToWSL(%q) = %q %v", picked, got, ok)
		}
	}

	if _, ok := uncToWSL(`C:\a.txt`); ok {
		t.Fatal("drive path treated as UNC")
	}
}

func TestFromWindowsPath(t *testing.T) {
	dir := t.TempDir()
	writeTool(t, dir, "wslpath", wslpathConvert)
	t.Setenv("PATH", dir)

	got, err := fromWindowsPath(context.Background(), `C:\Users\me\notes.txt`)
	if err != nil || got != "/mnt/c/Users/me/notes.txt" {
		t.Fatalf("path = %q %v", got, err)
	}
}

func TestFromWindowsPathUNC(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	got, err := fromWindowsPath(context.Background(), `\\wsl$\Ubuntu\home\me\a.txt`)
	if err != nil || got != "/home/me/a.txt" {
		t.Fatalf("path = %q %v", got, err)
	}
}

func TestPickWSLUsesWindows(t *testing.T) {
	dir := t.TempDir()
	writeTool(t, dir, "powershell.exe", powerShellPick)
	writeTool(t, dir, "wslpath", wslpathConvert)
	writeTool(t, dir, "zenity", "#!/bin/sh\nprintf '%s\\n' /tmp/linux.txt\n")
	t.Setenv("PATH", dir)
	t.Setenv("WSL_DISTRO_NAME", "Ubuntu")

	got, err := Pick(context.Background(), "Open")
	if err != nil || got != "/mnt/c/Users/me/notes.txt" {
		t.Fatalf("pick = %q %v", got, err)
	}
}
