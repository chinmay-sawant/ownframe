package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const recentProjectsMax = 10

func recentProjectsPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(base, "ownframe", "code-view", "recent-projects.json"), nil
}

func loadRecentProjects() []string {
	path, err := recentProjectsPath()
	if err != nil {
		return nil
	}

	data, err := os.ReadFile(path) //nolint:gosec // caller-owned config file under the user config dir.
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return nil
	}

	var roots []string

	if err := json.Unmarshal(data, &roots); err != nil {
		return nil
	}

	return cleanRecentProjects(roots)
}

func cleanRecentProjects(roots []string) []string {
	out := make([]string, 0, len(roots))
	seen := map[string]struct{}{}

	for _, raw := range roots {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}

		if trimmed == "sample" {
			if _, dup := seen["sample"]; dup {
				continue
			}

			seen["sample"] = struct{}{}
			out = append(out, "sample")

			if len(out) >= recentProjectsMax {
				break
			}

			continue
		}

		abs, err := filepath.Abs(trimmed)
		if err != nil {
			continue
		}

		abs = filepath.Clean(abs)

		if _, dup := seen[abs]; dup {
			continue
		}

		seen[abs] = struct{}{}
		out = append(out, abs)

		if len(out) >= recentProjectsMax {
			break
		}
	}

	return out
}

func rememberRecentProject(roots []string, root string) []string {
	trimmed := strings.TrimSpace(root)
	if trimmed == "sample" {
		out := []string{"sample"}

		for _, raw := range roots {
			if strings.TrimSpace(raw) == "sample" {
				continue
			}

			out = append(out, raw)

			if len(out) >= recentProjectsMax {
				break
			}
		}

		return out
	}

	abs, err := filepath.Abs(trimmed)
	if err != nil {
		return roots
	}

	abs = filepath.Clean(abs)
	out := []string{abs}

	for _, raw := range roots {
		if filepath.Clean(raw) == abs {
			continue
		}

		out = append(out, raw)

		if len(out) >= recentProjectsMax {
			break
		}
	}

	return out
}

func saveRecentProjects(roots []string) {
	path, err := recentProjectsPath()
	if err != nil {
		return
	}

	roots = cleanRecentProjects(roots)
	data, err := json.MarshalIndent(roots, "", "  ")
	if err != nil {
		return
	}

	data = append(data, '\n')

	dir := filepath.Dir(path)
	_ = os.MkdirAll(dir, sessionDirMode)

	tmp, err := os.CreateTemp(dir, ".recent-projects-*.tmp")
	if err != nil {
		return
	}

	name := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(name)

		return
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(name)

		return
	}

	_ = os.Rename(name, path)
}
