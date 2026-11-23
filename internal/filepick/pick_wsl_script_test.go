//go:build linux && !android

package filepick

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSecWindowScriptStatic asserts smart-quote titles never reach the script.
func TestSecWindowScriptStatic(t *testing.T) {
	t.Parallel()

	script := windowScript()
	for _, title := range []string{
		"it's; $(Remove-Item C:\\x)",
		"\u2019 ; calc ; \u2018",
		"\u201a ; calc ; \u201b",
	} {
		if strings.Contains(script, title) || strings.Contains(script, "calc") {
			t.Fatalf("script interpolates title %q: %q", title, script)
		}
	}
	if !strings.Contains(script, "$f.Title = $env:"+titleEnv+";") {
		t.Fatalf("script does not read title from env: %q", script)
	}
}

func TestSecWindowEnvForwardsTitle(t *testing.T) {
	t.Parallel()

	title := "\u2019 ; calc ; \u2018"
	got := windowEnv([]string{"A=1", "WSLENV=PATH/l", titleEnv + "=stale"}, title)
	want := map[string]bool{
		"A=1":                              true,
		titleEnv + "=" + title:             true,
		"WSLENV=PATH/l:" + titleEnv + "/u": true,
	}
	if len(got) != len(want) {
		t.Fatalf("env = %q", got)
	}
	for _, kv := range got {
		if !want[kv] {
			t.Fatalf("unexpected %q in %q", kv, got)
		}
	}
}

func TestSecPickWindowsPassesTitleViaEnv(t *testing.T) {
	dir := t.TempDir()
	out := dir + "/seen"
	writeTool(t, dir, "powershell.exe",
		"#!/bin/sh\nprintf '%s' \"$OWNFRAME_DIALOG_TITLE\" > "+out+"\nprintf '%s\\n' 'C:\\x'\n")
	writeTool(t, dir, "wslpath", wslpathConvert)
	t.Setenv("PATH", dir)

	title := "\u2019 ; calc ; \u2018"
	if _, err := pickWindows(context.Background(), title); err != nil {
		t.Fatal(err)
	}
	seen, err := os.ReadFile(filepath.Clean(out))
	if err != nil || string(seen) != title {
		t.Fatalf("env title = %q %v", seen, err)
	}
}
