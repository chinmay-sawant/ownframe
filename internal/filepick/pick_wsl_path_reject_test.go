//go:build linux && !android

package filepick

import (
	"context"
	"testing"
)

func TestFromWindowsPathRejectsUnsafeConvert(t *testing.T) {
	for _, out := range []string{
		"/etc/../secret",
		"/home/./me/notes.txt",
		"mnt/c/Users/me/notes.txt",
		"..",
	} {
		dir := t.TempDir()
		writeTool(t, dir, "wslpath", "#!/bin/sh\nprintf '%s\\n' '"+out+"'\n")
		t.Setenv("PATH", dir)

		if got, ok := fromWindowsPath(context.Background(), `C:\Users\me\notes.txt`); ok {
			t.Fatalf("fromWindowsPath accepted %q", got)
		}
	}
}
