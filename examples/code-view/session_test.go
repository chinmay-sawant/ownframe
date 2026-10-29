package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestSessionRoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "session.json")

	want := sessionState{
		Root:          "/work/repo",
		Mode:          "changes",
		Recents:       []string{"README.md", "backend/probe.go"},
		Tabs:          []string{"README.md", "backend/probe.go"},
		Active:        1,
		Tabs2:         []string{"web/theme.css"},
		Active2:       0,
		FocusGroup:    1,
		Split:         true,
		SplitRatio:    0.62,
		Details:       true,
		DevTools:      false,
		Wrap:          false,
		Preview:       true,
		DiffSplit:     true,
		HistoryRatio:  0.47,
		HistoryManual: true,
	}

	if err := saveSession(path, want); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := loadSession(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip:\n got %+v\nwant %+v", got, want)
	}

	if _, err := loadSession(""); err == nil {
		t.Fatal("load accepted an empty path")
	}

	if err := saveSession("", want); err == nil {
		t.Fatal("save accepted an empty path")
	}

	many := make([]string, sessionMaxPaths+5)
	for i := range many {
		many[i] = fmt.Sprintf("src/file-%02d.go", i)
	}

	if err := saveSession(path, sessionState{Recents: many, Tabs: many}); err != nil {
		t.Fatalf("save capped state: %v", err)
	}

	capped, err := loadSession(path)
	if err != nil {
		t.Fatalf("load capped state: %v", err)
	}

	if len(capped.Tabs) != sessionMaxPaths || len(capped.Recents) != sessionMaxPaths {
		t.Fatalf("caps: tabs = %d, recents = %d", len(capped.Tabs), len(capped.Recents))
	}

	if capped.Tabs[sessionMaxPaths-1] != many[sessionMaxPaths-1] {
		t.Fatalf("cap kept %q, want the first %d entries", capped.Tabs[sessionMaxPaths-1], sessionMaxPaths)
	}

	leftovers, err := filepath.Glob(filepath.Join(dir, ".session-*"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}

	if len(leftovers) != 0 {
		t.Fatalf("temporary files left behind: %v", leftovers)
	}
}

func TestSessionMissingIsZero(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing", "session.json")

	got, err := loadSession(path)
	if err != nil {
		t.Fatalf("missing file returned an error: %v", err)
	}

	if !reflect.DeepEqual(got, sessionState{}) {
		t.Fatalf("missing file state = %+v, want zero", got)
	}
}

func TestSessionCorrupt(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"broken": "{not json",
		"empty":  "",
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "session.json")
			if err := os.WriteFile(path, []byte(body), sessionFileMode); err != nil {
				t.Fatalf("write corrupt file: %v", err)
			}

			got, err := loadSession(path)
			if err == nil {
				t.Fatal("corrupt file loaded without an error")
			}

			if !strings.Contains(err.Error(), "session") || !strings.Contains(err.Error(), "decode") {
				t.Fatalf("error is not descriptive: %v", err)
			}

			if !reflect.DeepEqual(got, sessionState{}) {
				t.Fatalf("corrupt file state = %+v, want zero", got)
			}
		})
	}
}

func TestSessionSanitizes(t *testing.T) {
	t.Parallel()

	raw := sessionState{
		Mode: "changes",
		Recents: []string{
			" README.md ",
			"",
			"README.md",
			"/etc/passwd",
			"../secret",
			"backend/probe.go",
			"backend/../web/theme.css",
			`C:\Users\me\secret`,
		},
		Tabs: []string{
			"README.md",
			"README.md",
			"",
			"/abs.go",
			"a/../b.go",
			"web/theme.css",
			"config/services.yaml",
			"backend/probe.go",
		},
		Active:        99,
		Tabs2:         []string{"backend/server.go", "backend/server.go", `C:\Users\me\secret`},
		Active2:       -7,
		FocusGroup:    9,
		Split:         true,
		SplitRatio:    2,
		Details:       true,
		DevTools:      true,
		Wrap:          true,
		Preview:       true,
		DiffSplit:     true,
		HistoryRatio:  -3,
		HistoryManual: true,
	}

	want := sessionState{
		Root:          "/work/repo",
		Mode:          "changes",
		Recents:       []string{"README.md", "backend/probe.go"},
		Tabs:          []string{"README.md", "web/theme.css", "config/services.yaml", "backend/probe.go"},
		Active:        3,
		Tabs2:         []string{"backend/server.go"},
		Active2:       0,
		FocusGroup:    1,
		Split:         true,
		SplitRatio:    sessionSplitMax,
		Details:       true,
		DevTools:      true,
		Wrap:          true,
		Preview:       true,
		DiffSplit:     true,
		HistoryRatio:  sessionHistoryMin,
		HistoryManual: true,
		Zoom:          1,
	}

	if got := raw.sanitize("/work/repo"); !reflect.DeepEqual(got, want) {
		t.Fatalf("sanitize:\n got %+v\nwant %+v", got, want)
	}

	low := sessionState{
		Root:         "  kept  ",
		Active:       -2,
		Active2:      99,
		FocusGroup:   -3,
		SplitRatio:   -1,
		HistoryRatio: 9,
	}.sanitize("ignored")

	if low.Root != "kept" || low.Active != 0 || low.Active2 != 0 || low.FocusGroup != 0 {
		t.Fatalf("indexes or root not clamped: %+v", low)
	}

	if low.SplitRatio != sessionSplitMin || low.HistoryRatio != sessionHistoryMax {
		t.Fatalf("ratios = %v and %v", low.SplitRatio, low.HistoryRatio)
	}

	if low.Zoom != 1 {
		t.Fatalf("zoom = %v, want 1", low.Zoom)
	}

	many := make([]string, sessionMaxPaths+3)
	for i := range many {
		many[i] = fmt.Sprintf("pkg/file-%02d.go", i)
	}

	capped := sessionState{Tabs: many, Recents: many}.sanitize("root")
	if len(capped.Tabs) != sessionMaxPaths || len(capped.Recents) != sessionMaxPaths {
		t.Fatalf("sanitize caps: tabs = %d, recents = %d", len(capped.Tabs), len(capped.Recents))
	}
}

func TestSessionDefaultPathIsPerRoot(t *testing.T) {
	t.Parallel()

	firstRoot := t.TempDir()
	secondRoot := t.TempDir()

	first, err := defaultSessionPath(firstRoot)
	if err != nil {
		t.Fatalf("first path: %v", err)
	}

	again, err := defaultSessionPath(firstRoot)
	if err != nil {
		t.Fatalf("repeat path: %v", err)
	}

	second, err := defaultSessionPath(secondRoot)
	if err != nil {
		t.Fatalf("second path: %v", err)
	}

	if first != again {
		t.Fatalf("same root changed path: %q then %q", first, again)
	}

	if first == second {
		t.Fatalf("two roots share one session file: %q", first)
	}

	base, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("config dir: %v", err)
	}

	if want := filepath.Join(base, "ownframe", "code-view"); filepath.Dir(first) != want {
		t.Fatalf("session dir = %q, want %q", filepath.Dir(first), want)
	}

	abs, err := filepath.Abs(firstRoot)
	if err != nil {
		t.Fatalf("absolute root: %v", err)
	}

	sum := sha256.Sum256([]byte(filepath.Clean(abs)))
	wantName := "session-" + hex.EncodeToString(sum[:sessionHashBytes]) + ".json"

	if filepath.Base(first) != wantName {
		t.Fatalf("session name = %q, want %q", filepath.Base(first), wantName)
	}
}

func TestSessionFileMode(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("Windows does not report Unix permission bits")
	}

	dir := filepath.Join(t.TempDir(), "state")
	path := filepath.Join(dir, "session.json")

	if err := saveSession(path, sessionState{Root: "/work/repo"}); err != nil {
		t.Fatalf("save: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat file: %v", err)
	}

	if got := info.Mode().Perm(); got != sessionFileMode {
		t.Fatalf("file mode = %#o, want %#o", got, sessionFileMode)
	}

	info, err = os.Stat(dir)
	if err != nil {
		t.Fatalf("stat dir: %v", err)
	}

	if got := info.Mode().Perm(); got != sessionDirMode {
		t.Fatalf("dir mode = %#o, want %#o", got, sessionDirMode)
	}

	if err := os.Chmod(path, 0o400); err != nil {
		t.Fatalf("chmod: %v", err)
	}

	if err := saveSession(path, sessionState{Root: "/work/repo", Wrap: true}); err != nil {
		t.Fatalf("overwrite: %v", err)
	}

	info, err = os.Stat(path)
	if err != nil {
		t.Fatalf("stat replaced file: %v", err)
	}

	if got := info.Mode().Perm(); got != sessionFileMode {
		t.Fatalf("replaced file mode = %#o, want %#o", got, sessionFileMode)
	}
}
