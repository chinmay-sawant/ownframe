package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	pathpkg "path"
	"path/filepath"
	"strings"
)

const (
	sessionMaxPaths   = 32
	sessionHashBytes  = 16
	sessionDirMode    = 0o700
	sessionFileMode   = 0o600
	sessionSplitMin   = 0.2
	sessionSplitMax   = 0.8
	sessionGroupCount = 2
	sessionHistoryMin = historyMinRatio
	sessionHistoryMax = historyMaxRatio
)

// sessionState is the layout and open-document state one workspace restores
// after a restart. Paths are workspace-relative, the two tab lists belong to
// the two editor groups, and the ratios are fractions of the editor body.
type sessionState struct {
	Root          string               `json:"root"`
	Mode          string               `json:"mode"`
	Recents       []string             `json:"recents"`
	Tabs          []string             `json:"tabs"`
	Active        int                  `json:"active"`
	Tabs2         []string             `json:"tabs2"`
	Active2       int                  `json:"active2"`
	FocusGroup    int                  `json:"focusGroup"`
	Split         bool                 `json:"split"`
	SplitRatio    float64              `json:"splitRatio"`
	Details       bool                 `json:"details"`
	DevTools      bool                 `json:"devTools"`
	Wrap          bool                 `json:"wrap"`
	Preview       bool                 `json:"preview"`
	DiffSplit     bool                 `json:"diffSplit"`
	HistoryRatio  float64              `json:"historyRatio"`
	HistoryManual bool                 `json:"historyManual"`
	BranchLogs    map[string]commitLog `json:"branchLogs,omitempty"`
}

// defaultSessionPath returns the session file for root under the user config
// directory. The absolute root is hashed so two checkouts keep separate
// sessions while one checkout keeps the same name across runs. It does not
// create anything; saveSession makes the directory with owner-only mode.
func defaultSessionPath(root string) (string, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("session: absolute root: %w", err)
	}

	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("session: config dir: %w", err)
	}

	sum := sha256.Sum256([]byte(filepath.Clean(abs)))
	name := "session-" + hex.EncodeToString(sum[:sessionHashBytes]) + ".json"

	return filepath.Join(base, "ownframe", "code-view", name), nil
}

// loadSession reads one state file. A missing file is a first run, so it
// returns the zero state without an error. A damaged file also returns the
// zero state, with the decode error wrapped for the caller to report.
func loadSession(path string) (sessionState, error) {
	var state sessionState

	if strings.TrimSpace(path) == "" {
		return state, errors.New("session: empty path")
	}

	data, err := os.ReadFile(path) //nolint:gosec // the caller names the session file, either under the user config dir or in a test temp dir.
	if errors.Is(err, os.ErrNotExist) {
		return state, nil
	}

	if err != nil {
		return state, fmt.Errorf("session: read %s: %w", path, err)
	}

	if err := json.Unmarshal(data, &state); err != nil {
		return sessionState{}, fmt.Errorf("session: decode %s: %w", path, err)
	}

	state.Recents = capSessionPaths(state.Recents)
	state.Tabs = capSessionPaths(state.Tabs)
	state.Tabs2 = capSessionPaths(state.Tabs2)

	return state, nil
}

// saveSession writes one state file through a temporary file in the same
// directory, then renames it into place so a crash cannot leave half a file.
// The parent directory and the file are owner-only.
func saveSession(path string, state sessionState) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("session: empty path")
	}

	state.Recents = capSessionPaths(state.Recents)
	state.Tabs = capSessionPaths(state.Tabs)
	state.Tabs2 = capSessionPaths(state.Tabs2)

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("session: encode %s: %w", path, err)
	}

	data = append(data, '\n')

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, sessionDirMode); err != nil {
		return fmt.Errorf("session: create %s: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, ".session-*.tmp")
	if err != nil {
		return fmt.Errorf("session: temp in %s: %w", dir, err)
	}

	name := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(name)

		return fmt.Errorf("session: write %s: %w", name, err)
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(name)

		return fmt.Errorf("session: close %s: %w", name, err)
	}

	if err := os.Rename(name, path); err != nil {
		_ = os.Remove(name)

		return fmt.Errorf("session: replace %s: %w", path, err)
	}

	return nil
}

// sanitize returns a state that is safe to apply: relative unique paths only,
// indices inside their tab lists, a valid focus group, ratios inside the pane
// limits, and the current root when the saved state named none.
func (s sessionState) sanitize(root string) sessionState {
	s.Root = strings.TrimSpace(s.Root)
	if s.Root == "" {
		s.Root = strings.TrimSpace(root)
	}

	s.Recents = cleanSessionPaths(s.Recents)
	s.Tabs = cleanSessionPaths(s.Tabs)
	s.Tabs2 = cleanSessionPaths(s.Tabs2)

	s.Active = clampSessionIndex(s.Active, len(s.Tabs))
	s.Active2 = clampSessionIndex(s.Active2, len(s.Tabs2))
	s.FocusGroup = clampSessionIndex(s.FocusGroup, sessionGroupCount)
	s.SplitRatio = clampSessionRatio(s.SplitRatio, sessionSplitMin, sessionSplitMax)
	s.HistoryRatio = clampSessionRatio(s.HistoryRatio, sessionHistoryMin, sessionHistoryMax)

	return s
}

// cleanSessionPaths trims, normalizes, and de-duplicates paths, drops entries
// that are absolute or climb out of the workspace, and keeps the first 32.
func cleanSessionPaths(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}

	out := make([]string, 0, min(len(paths), sessionMaxPaths))
	seen := make(map[string]struct{}, min(len(paths), sessionMaxPaths))

	for _, raw := range paths {
		path, ok := cleanSessionPath(raw)
		if !ok {
			continue
		}

		if _, dup := seen[path]; dup {
			continue
		}

		seen[path] = struct{}{}
		out = append(out, path)

		if len(out) == sessionMaxPaths {
			break
		}
	}

	if len(out) == 0 {
		return nil
	}

	return out
}

// cleanSessionPath turns one saved path into a workspace-relative slash path.
// It rejects absolute paths, Windows volume paths, and any parent segment.
func cleanSessionPath(raw string) (string, bool) {
	path := strings.ReplaceAll(strings.TrimSpace(raw), "\\", "/")
	if path == "" || strings.ContainsRune(path, '\x00') {
		return "", false
	}

	if pathpkg.IsAbs(path) || (len(path) >= 2 && path[1] == ':') {
		return "", false
	}

	for _, part := range strings.Split(path, "/") {
		if part == ".." {
			return "", false
		}
	}

	path = pathpkg.Clean(path)
	if path == "." || path == ".." || strings.HasPrefix(path, "../") {
		return "", false
	}

	return path, true
}

// capSessionPaths keeps a loaded or saved list inside the persistence budget
// without touching the caller's entries.
func capSessionPaths(paths []string) []string {
	if len(paths) > sessionMaxPaths {
		return paths[:sessionMaxPaths]
	}

	return paths
}

// clampSessionIndex keeps an index inside a list of length items, with zero
// as the empty-list index.
func clampSessionIndex(index, length int) int {
	if index < 0 || length <= 0 {
		return 0
	}

	if index >= length {
		return length - 1
	}

	return index
}

// clampSessionRatio keeps a pane fraction inside its usable range.
func clampSessionRatio(value, low, high float64) float64 {
	if value < low {
		return low
	}

	if value > high {
		return high
	}

	return value
}
